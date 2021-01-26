/*
SPDX-FileCopyrightText: The Kubernetes Authors.

SPDX-License-Identifier: Apache-2.0
*/

package resources

import (
	"context"
	"reflect"
	"sync"
	"time"

	"github.com/gardener/controller-manager-library/pkg/ctxutil"
	"github.com/gardener/controller-manager-library/pkg/kutil"
	"github.com/gardener/controller-manager-library/pkg/logger"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

type GenericInformer interface {
	// Selected Methods from cache.SharedInformer

	AddEventHandler(handler cache.ResourceEventHandler)
	HasSynced() bool
	LastSyncResourceVersion() string
	Run()

	//Informer() cache.SharedIndexInformer
	Lister() Lister
	Context() context.Context
}

type event struct {
	New interface{}
	Old interface{}
}

type wrappedHandler struct {
	handler   cache.ResourceEventHandler
	processor *kutil.Processor
}

func newWrappedHandler(ctx context.Context, handler cache.ResourceEventHandler) *wrappedHandler {
	ret := &wrappedHandler{
		handler: handler,
	}
	ret.processor = kutil.NewProcessor(ctx, ret.handle, 100)
	return ret
}

func (w *wrappedHandler) handle(e interface{}) {
	evt := e.(*event)
	if evt.New == nil {
		w.handler.OnDelete(evt.Old)
	} else if evt.Old == nil {
		w.handler.OnAdd(evt.New)
	} else {
		w.handler.OnUpdate(evt.Old, evt.New)
	}
}

func (w *wrappedHandler) OnAdd(obj interface{}) {
	w.processor.Add(&event{obj, nil})
}
func (w *wrappedHandler) OnUpdate(oldObj, newObj interface{}) {
	w.processor.Add(&event{newObj, oldObj})
}
func (w *wrappedHandler) OnDelete(obj interface{}) {
	w.processor.Add(&event{nil, obj})
}

type genericHandler struct {
	lock     sync.RWMutex
	ctx      context.Context
	handlers map[cache.ResourceEventHandler]*wrappedHandler
}

func newGenericHandlerr(ctx context.Context) *genericHandler {
	ret := &genericHandler{handlers: map[cache.ResourceEventHandler]*wrappedHandler{}, ctx: ctx}
	return ret
}

func (w *genericHandler) StopAndWait() {
	w.lock.Lock()
	defer w.lock.Unlock()
	for _, p := range w.handlers {
		p.processor.Stop()
	}
	for h, p := range w.handlers {
		delete(w.handlers, h)
		p.processor.Wait()
	}
}

func (w *genericHandler) Size() int {
	w.lock.Lock()
	defer w.lock.Unlock()
	return len(w.handlers)
}

func (w *genericHandler) AddEventHandler(h cache.ResourceEventHandler) {
	w.lock.Lock()
	defer w.lock.Unlock()
	if _, ok := w.handlers[h]; !ok {
		w.handlers[h] = newWrappedHandler(w.ctx, h)
	}
}

func (w *genericHandler) RemoveEventHandler(h cache.ResourceEventHandler) {
	var p *wrappedHandler
	var ok bool

	func() {
		w.lock.Lock()
		defer w.lock.Unlock()

		if p, ok = w.handlers[h]; ok {
			delete(w.handlers, h)
		}
	}()

	if p != nil {
		p.processor.Stop()
		p.processor.Wait()
	}
}

func (w *genericHandler) OnAdd(obj interface{}) {
	w.lock.RLock()
	defer w.lock.RUnlock()
	for _, h := range w.handlers {
		h.OnAdd(obj)
	}
}
func (w *genericHandler) OnUpdate(oldObj, newObj interface{}) {
	w.lock.RLock()
	defer w.lock.RUnlock()
	for _, h := range w.handlers {
		h.OnUpdate(oldObj, newObj)
	}
}
func (w *genericHandler) OnDelete(obj interface{}) {
	w.lock.RLock()
	defer w.lock.RUnlock()
	for _, h := range w.handlers {
		h.OnDelete(obj)
	}
}

type genericInformer struct {
	lock        sync.Mutex
	lw          listWatchFactory
	namespace   string
	optionsFunc TweakListOptionsFunc

	cache.SharedIndexInformer
	context context.Context
	lister  Lister
	generic *genericHandler
}

func newGenericInformer(ctx context.Context, lw listWatchFactory, namespace string, optionsFunc TweakListOptionsFunc) (*genericInformer, error) {
	ret := &genericInformer{
		lw:          lw,
		namespace:   namespace,
		optionsFunc: optionsFunc,
		context:     ctx,
	}
	err := ret.new()
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (f *genericInformer) new() error {
	logger.Infof("new generic informer for %s (%s) %s (%d seconds)", f.lw.ElemType(), f.lw.GroupVersionKind(), f.lw.ListType(), f.lw.Resync()/time.Second)

	ctx := ctxutil.CancelContext(f.context)
	listWatch, err := f.lw.CreateListWatch(ctx, f.namespace, f.optionsFunc)
	if err != nil {
		return err
	}
	f.SharedIndexInformer = cache.NewSharedIndexInformer(listWatch, f.lw.ExampleObject(), resyncPeriod(f.lw.Resync())(),
		cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	f.generic = newGenericHandlerr(ctx)
	return nil
}

func (f *genericInformer) Run() {
	f.SharedIndexInformer.AddEventHandler(f.generic)
	f.SharedIndexInformer.Run(f.context.Done())
}

func (w *genericInformer) AddEventHandler(h cache.ResourceEventHandler) {
	w.lock.Lock()
	defer w.lock.Unlock()

	if w.SharedIndexInformer == nil {
		w.new()
		go w.Run()
		cache.WaitForCacheSync(w.context.Done(), w.HasSynced)
	}
	w.generic.AddEventHandler(h)
}

func (w *genericInformer) RemoveEventHandler(h cache.ResourceEventHandler) {
	w.lock.Lock()
	defer w.lock.Unlock()
	w.generic.RemoveEventHandler(h)
	if w.lister == nil && w.generic.Size() == 0 {
		w.generic.StopAndWait()
		w.SharedIndexInformer = nil
		w.generic = nil
	}
}

func (w *genericInformer) IsUsed() bool {
	w.lock.Lock()
	defer w.lock.Unlock()
	return w.lister == nil && w.generic.Size() == 0
}

func (f *genericInformer) Informer() cache.SharedIndexInformer {
	return f.SharedIndexInformer
}

func (f *genericInformer) Lister() Lister {
	f.lock.Lock()
	defer f.lock.Unlock()
	if f.lister == nil {
		if f.SharedIndexInformer == nil {
			f.new()
		}
		f.lister = NewLister(f.SharedIndexInformer.GetIndexer(), f.lw.Info())
	}
	return f.lister
}

func (f *genericInformer) Context() context.Context {
	return f.context
}

func (f *genericInformer) HasSynced() bool {
	select {
	case <-f.context.Done():
		return true
	default:
		return f.SharedIndexInformer.HasSynced()
	}
}

// SharedInformerFactory provides shared informers for resources in all known
// API group versions.
type SharedInformerFactory interface {
	Structured() GenericFilteredInformerFactory
	Unstructured() GenericFilteredInformerFactory
	MinimalObject() GenericFilteredInformerFactory

	InformerForObject(obj runtime.Object) (GenericInformer, error)
	FilteredInformerForObject(obj runtime.Object, namespace string, optionsFunc TweakListOptionsFunc) (GenericInformer, error)

	InformerFor(gvk schema.GroupVersionKind) (GenericInformer, error)
	FilteredInformerFor(gvk schema.GroupVersionKind, namespace string, optionsFunc TweakListOptionsFunc) (GenericInformer, error)

	UnstructuredInformerFor(gvk schema.GroupVersionKind) (GenericInformer, error)
	FilteredUnstructuredInformerFor(gvk schema.GroupVersionKind, namespace string, optionsFunc TweakListOptionsFunc) (GenericInformer, error)

	MinimalObjectInformerFor(gvk schema.GroupVersionKind) (GenericInformer, error)
	FilteredMinimalObjectInformerFor(gvk schema.GroupVersionKind, namespace string, optionsFunc TweakListOptionsFunc) (GenericInformer, error)

	Start()
	WaitForCacheSync()
}

type GenericInformerFactory interface {
	InformerFor(gvk schema.GroupVersionKind) (GenericInformer, error)
	Start()
	WaitForCacheSync()
}

type GenericFilteredInformerFactory interface {
	GenericInformerFactory
	FilteredInformerFor(gvk schema.GroupVersionKind, namespace string, optionsFunc TweakListOptionsFunc) (GenericInformer, error)
	LookupInformerFor(gvk schema.GroupVersionKind, namespace string) (GenericInformer, error)
}

///////////////////////////////////////////////////////////////////////////////
//  informer factory

type sharedInformerFactory struct {
	context       *resourceContext
	structured    *sharedFilteredInformerFactory
	unstructured  *sharedFilteredInformerFactory
	minimalObject *sharedFilteredInformerFactory
}

func newSharedInformerFactory(rctx *resourceContext, defaultResync time.Duration) *sharedInformerFactory {
	return &sharedInformerFactory{
		context:       rctx,
		structured:    newSharedFilteredInformerFactory(rctx, defaultResync, newStructuredListWatchFactory),
		unstructured:  newSharedFilteredInformerFactory(rctx, defaultResync, newUnstructuredListWatchFactory),
		minimalObject: newSharedFilteredInformerFactory(rctx, defaultResync, newMinimalObjectListWatchFactory),
	}
}

func (f *sharedInformerFactory) Structured() GenericFilteredInformerFactory {
	return f.structured
}

func (f *sharedInformerFactory) Unstructured() GenericFilteredInformerFactory {
	return f.unstructured
}

func (f *sharedInformerFactory) MinimalObject() GenericFilteredInformerFactory {
	return f.minimalObject
}

// Start initializes all requested informers.
func (f *sharedInformerFactory) Start() {
	f.structured.Start()
	f.unstructured.Start()
	f.minimalObject.Start()
}

func (f *sharedInformerFactory) WaitForCacheSync() {
	f.structured.WaitForCacheSync()
	f.unstructured.WaitForCacheSync()
	f.minimalObject.WaitForCacheSync()
}

func (f *sharedInformerFactory) UnstructuredInformerFor(gvk schema.GroupVersionKind) (GenericInformer, error) {
	return f.FilteredUnstructuredInformerFor(gvk, "", nil)
}

func (f *sharedInformerFactory) FilteredUnstructuredInformerFor(gvk schema.GroupVersionKind, namespace string, optionsFunc TweakListOptionsFunc) (GenericInformer, error) {
	return f.unstructured.informerFor(gvk, namespace, optionsFunc)
}

func (f *sharedInformerFactory) MinimalObjectInformerFor(gvk schema.GroupVersionKind) (GenericInformer, error) {
	return f.FilteredMinimalObjectInformerFor(gvk, "", nil)
}

func (f *sharedInformerFactory) FilteredMinimalObjectInformerFor(gvk schema.GroupVersionKind, namespace string, optionsFunc TweakListOptionsFunc) (GenericInformer, error) {
	return f.minimalObject.informerFor(gvk, namespace, optionsFunc)
}

func (f *sharedInformerFactory) InformerFor(gvk schema.GroupVersionKind) (GenericInformer, error) {
	return f.FilteredInformerFor(gvk, "", nil)
}

func (f *sharedInformerFactory) FilteredInformerFor(gvk schema.GroupVersionKind, namespace string, optionsFunc TweakListOptionsFunc) (GenericInformer, error) {
	return f.structured.informerFor(gvk, namespace, optionsFunc)
}

func (f *sharedInformerFactory) LookupInformerFor(gvk schema.GroupVersionKind, namespace string) (GenericInformer, error) {
	return f.structured.lookupInformerFor(gvk, namespace)
}

func (f *sharedInformerFactory) InformerForObject(obj runtime.Object) (GenericInformer, error) {
	return f.FilteredInformerForObject(obj, "", nil)
}

func (f *sharedInformerFactory) FilteredInformerForObject(obj runtime.Object, namespace string, optionsFunc TweakListOptionsFunc) (GenericInformer, error) {
	informerType := reflect.TypeOf(obj)
	for informerType.Kind() == reflect.Ptr {
		informerType = informerType.Elem()
	}

	gvk, err := f.context.GetGVK(obj)
	if err != nil {
		return nil, err
	}
	return f.FilteredInformerFor(gvk, namespace, optionsFunc)
}

///////////////////////////////////////////////////////////////////////////////
// Shared Filtered Informer Factory

type sharedFilteredInformerFactory struct {
	lock sync.Mutex

	context                    *resourceContext
	defaultResync              time.Duration
	filters                    map[string]*genericInformerFactory
	newListWatchFactoryFactory newListWatchFactoryFactory
}

func newSharedFilteredInformerFactory(rctx *resourceContext, defaultResync time.Duration, ff newListWatchFactoryFactory) *sharedFilteredInformerFactory {
	return &sharedFilteredInformerFactory{
		context:       rctx,
		defaultResync: defaultResync,

		filters:                    make(map[string]*genericInformerFactory),
		newListWatchFactoryFactory: ff,
	}
}

// Start initializes all requested informers.
func (f *sharedFilteredInformerFactory) Start() {
	for _, i := range f.filters {
		i.Start()
	}
}

func (f *sharedFilteredInformerFactory) WaitForCacheSync() {
	for _, i := range f.filters {
		i.WaitForCacheSync()
	}
}

func (f *sharedFilteredInformerFactory) getFactory(namespace string, optionsFunc TweakListOptionsFunc) *genericInformerFactory {
	key := namespace
	if optionsFunc != nil {
		opts := metav1.ListOptions{}
		optionsFunc(&opts)
		key = namespace + opts.String()
	}

	f.lock.Lock()
	defer f.lock.Unlock()

	factory, exists := f.filters[key]
	if !exists {
		factory = newGenericInformerFactory(f.context, f.defaultResync, namespace, optionsFunc)
		f.filters[key] = factory
	}
	return factory
}

func (f *sharedFilteredInformerFactory) queryFactory(namespace string) *genericInformerFactory {
	f.lock.Lock()
	defer f.lock.Unlock()

	factory, _ := f.filters[namespace]
	return factory
}

func (f *sharedFilteredInformerFactory) informerFor(gvk schema.GroupVersionKind, namespace string, optionsFunc TweakListOptionsFunc) (GenericInformer, error) {
	lwFactory, err := f.newListWatchFactoryFactory(f.context, gvk)
	if err != nil {
		return nil, err
	}
	return f.getFactory(namespace, optionsFunc).informerFor(lwFactory)
}

func (f *sharedFilteredInformerFactory) lookupInformerFor(gvk schema.GroupVersionKind, namespace string) (GenericInformer, error) {
	fac := f.queryFactory("")
	if fac != nil {
		i := fac.queryInformerFor(gvk)
		if i != nil {
			return i, nil
		}
	}
	if namespace != "" {
		fac := f.queryFactory(namespace)
		if fac != nil {
			i := fac.queryInformerFor(gvk)
			if i != nil {
				return i, nil
			}
			lwFactory, err := f.newListWatchFactoryFactory(f.context, gvk)
			if err != nil {
				return nil, err
			}
			return fac.informerFor(lwFactory)
		}
	}
	lwFactory, err := f.newListWatchFactoryFactory(f.context, gvk)
	if err != nil {
		return nil, err
	}
	return f.getFactory("", nil).informerFor(lwFactory)
}

func (f *sharedFilteredInformerFactory) InformerFor(gvk schema.GroupVersionKind) (GenericInformer, error) {
	return f.FilteredInformerFor(gvk, "", nil)
}

func (f *sharedFilteredInformerFactory) FilteredInformerFor(gvk schema.GroupVersionKind, namespace string, optionsFunc TweakListOptionsFunc) (GenericInformer, error) {
	return f.informerFor(gvk, namespace, optionsFunc)
}

func (f *sharedFilteredInformerFactory) LookupInformerFor(gvk schema.GroupVersionKind, namespace string) (GenericInformer, error) {
	return f.lookupInformerFor(gvk, namespace)
}

////////////////////////////////////////////////////////////////////////////////
// Watch

type watchWrapper struct {
	ctx        context.Context
	orig       watch.Interface
	origChan   <-chan watch.Event
	resultChan chan watch.Event
}

func NewWatchWrapper(ctx context.Context, orig watch.Interface) watch.Interface {
	logger.Infof("*************** new wrapper ********************")
	w := &watchWrapper{ctx, orig, orig.ResultChan(), make(chan watch.Event)}
	go w.Run()
	return w
}

func (w *watchWrapper) Stop() {
	w.orig.Stop()
}

func (w *watchWrapper) ResultChan() <-chan watch.Event {
	return w.resultChan
}
func (w *watchWrapper) Run() {
loop:
	for {
		select {
		case <-w.ctx.Done():
			break loop
		case e, ok := <-w.origChan:
			if !ok {
				logger.Infof("watch aborted")
				break loop
			} else {
				logger.Infof("WATCH: %#v\n", e)
				w.resultChan <- e
			}
		}
	}
	logger.Infof("stop wrapper ***************")
	close(w.resultChan)
}

var _ watch.Interface = &watchWrapper{}
