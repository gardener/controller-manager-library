/*
 * SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package controller

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	resourceserrors "github.com/gardener/controller-manager-library/pkg/resources/errors"
	resourcesv1alpha1 "github.com/gardener/gardener/pkg/apis/resources/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/set"

	"github.com/gardener/controller-manager-library/pkg/config"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/cluster"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/controller/lease"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/controller/mappings"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/controller/reconcile"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/extension"
	"github.com/gardener/controller-manager-library/pkg/ctxutil"
	"github.com/gardener/controller-manager-library/pkg/logger"
	"github.com/gardener/controller-manager-library/pkg/resources"
	"github.com/gardener/controller-manager-library/pkg/resources/apiextensions"
	"github.com/gardener/controller-manager-library/pkg/utils"
)

const A_MAINTAINER = "crds.gardener.cloud/maintainer"

type EventRecorder interface {
	// The resulting event will be created in the same namespace as the reference object.
	// Event(object runtime.ObjectData, eventtype, reason, message string)

	// Eventf is just like Event, but with Sprintf for the message field.
	// Eventf(object runtime.ObjectData, eventtype, reason, messageFmt string, args ...interface{})

	// AnnotatedEventf is just like eventf, but with annotations attached
	// AnnotatedEventf(object runtime.ObjectData, annotations map[string]string, eventtype, reason, messageFmt string, args ...interface{})
}

type ReconcilationElementSpec interface {
	String() string
}

var _ ReconcilationElementSpec = schema.GroupKind{}
var _ ReconcilationElementSpec = utils.Matcher(nil)

type _ReconcilationKey struct {
	key        ReconcilationElementSpec
	cluster    string
	reconciler string
}

type _Reconcilations map[_ReconcilationKey]string

func (this _Reconcilations) Get(cluster resources.Cluster, gk schema.GroupKind) utils.StringSet {
	reconcilers := utils.StringSet{}
	cluster_name := cluster.GetName()
	for k := range this {
		if k.cluster == cluster_name && k.key == gk {
			reconcilers.Add(k.reconciler)
		}
	}
	return reconcilers
}

type ReadyFlag struct {
	lock    sync.Mutex
	isready bool
}

func (this *ReadyFlag) WhenReady() {
	this.lock.Lock()
	defer this.lock.Unlock()
}

func (this *ReadyFlag) IsReady() bool {
	return this.isready
}

func (this *ReadyFlag) ready() {
	this.isready = true
	this.lock.Unlock()
}

func (this *ReadyFlag) start() {
	this.lock.Lock()
}

type watchContext struct {
	cluster    cluster.Interface
	controller Interface
}

func newWatchContext(controller Interface, cluster cluster.Interface) WatchContext {
	return watchContext{cluster, controller}
}

func (this watchContext) Cluster() cluster.Interface {
	return this.cluster
}

func (this watchContext) GetCluster(name string) cluster.Interface {
	return this.controller.GetCluster(name)
}

func (this watchContext) Name() string {
	return this.controller.GetName()
}

func (this watchContext) Namespace() string {
	return this.controller.GetEnvironment().Namespace()
}

func (this watchContext) GetOptionSource(name string) (config.OptionSource, error) {
	return this.controller.GetOptionSource(name)
}

func (this watchContext) GetOption(name string) (*config.ArbitraryOption, error) {
	return this.controller.GetOption(name)
}

func (this watchContext) GetBoolOption(name string) (bool, error) {
	return this.controller.GetBoolOption(name)
}

func (this watchContext) GetStringOption(name string) (string, error) {
	return this.controller.GetStringOption(name)
}

func (this watchContext) GetStringArrayOption(name string) ([]string, error) {
	return this.controller.GetStringArrayOption(name)
}

func (this watchContext) GetIntOption(name string) (int, error) {
	return this.controller.GetIntOption(name)
}

func (this watchContext) GetDurationOption(name string) (time.Duration, error) {
	return this.controller.GetDurationOption(name)
}

type watchDef struct {
	WatchResourceDef
	PoolName   string
	Reconciler string
}

func (this *watchDef) TweakListOptions(opts *metav1.ListOptions) {
	for _, t := range this.WatchResourceDef.Tweaker {
		t(opts)
	}
}

type controller struct {
	extension.ElementBase
	record.EventRecorder

	extension.SharedAttributes

	ready           ReadyFlag
	definition      Definition
	env             Environment
	cluster         cluster.Interface
	clusters        cluster.Clusters
	filters         []ResourceFilter
	owning          WatchResource
	mainresc        *watchDef
	watches         map[string][]*watchDef
	reconcilers     map[string]reconcile.Interface
	reconcilerNames map[reconcile.Interface]string
	mappings        _Reconcilations
	syncRequests    *SyncRequests
	finalizer       Finalizer

	options  *ControllerConfig
	handlers map[string]*ClusterHandler

	pools map[string]*pool

	lock   sync.Mutex
	leases map[string]map[string]bool
}

func Filter(_ ResourceKey, _ resources.Object) bool {
	return true
}

func NewController(env Environment, def Definition, cmp mappings.Definition) (*controller, error) {
	options := env.GetConfig().GetSource(def.Name()).(*ControllerConfig)

	this := &controller{
		definition: def,
		options:    options,
		env:        env,

		owning:  def.MainWatchResource(),
		filters: def.ResourceFilters(),

		handlers:        map[string]*ClusterHandler{},
		watches:         map[string][]*watchDef{},
		pools:           map[string]*pool{},
		reconcilers:     map[string]reconcile.Interface{},
		leases:          map[string]map[string]bool{},
		reconcilerNames: map[reconcile.Interface]string{},
		mappings:        _Reconcilations{},
		finalizer:       NewDefaultFinalizer(def.FinalizerName()),
	}

	this.syncRequests = NewSyncRequests(this)

	ctx := ctxutil.WaitGroupContext(env.GetContext(), "controller ", def.Name())
	this.ElementBase = extension.NewElementBase(ctx, ctx_controller, this, def.Name(), CONTROLLER_SET_PREFIX, options)
	this.SharedAttributes = extension.NewSharedAttributes(this.ElementBase)
	this.ready.start()

	required := cluster.Canonical(def.RequiredClusters())
	clusters, err := mappings.MapClusters(env.GetClusters(), cmp, required...)
	if err != nil {
		return nil, err
	}
	this.Infof("  using clusters %+v: %s (selected from %s)", required, clusters, env.GetClusters())
	if def.Scheme() != nil {
		if def.Scheme() != resources.DefaultScheme() {
			this.Infof("  using dedicated scheme for clusters")
		}
		clusters, err = clusters.WithScheme(def.Scheme())
		if err != nil {
			return nil, err
		}
	}
	this.clusters = clusters
	this.cluster = clusters.GetCluster(required[0])
	this.EventRecorder = this.cluster.Resources()
	this.mainresc = &watchDef{
		WatchResourceDef: this.owning.WatchResourceDef(newWatchContext(this, this.cluster)),
		PoolName:         DEFAULT_POOL,
		Reconciler:       DEFAULT_RECONCILER,
	}

	err = this.deployCRDS()
	if err != nil {
		return nil, err
	}

	for n, t := range def.Reconcilers() {
		this.Infof("creating reconciler %q", n)
		reconciler, err := t(this)
		if err != nil {
			return nil, fmt.Errorf("creating reconciler %s failed: %s", n, err)
		}
		this.reconcilers[n] = reconciler
		this.reconcilerNames[reconciler] = n
	}

	this.Infof("configure reconciler watches...")
	for cname, watches := range this.definition.Watches() {
		cluster := this.GetCluster(cname)
		if cluster == nil {
			return nil, fmt.Errorf("cluster %q not found for resource definitions", cname)
		}
		this.Infof("  for cluster %s", cluster)
		wctx := newWatchContext(this, cluster)
		for _, w := range watches {
			def := &watchDef{
				WatchResourceDef: w.WatchResourceDef(wctx),
				PoolName:         w.PoolName(),
				Reconciler:       w.Reconciler(),
			}
			key := def.Key
			if key != nil {
				if w.String() != key.String() {
					this.Infof("    %s mapped to %s", w, key)
				} else {
					this.Infof("    %s", key)
				}
				ok, err := this.addReconciler(cname, key.GroupKind(), w.PoolName(), w.Reconciler())
				if err != nil {
					this.Errorf("GOT error: %s", err)
					return nil, err
				}
				if ok {
					this.watches[cname] = append(this.watches[cname], def)
				} else {
					this.Infof("    omitted for reconciler", w.Reconciler())
				}
			} else {
				this.Infof("    no resource for %s", w)
			}
		}
	}

	_, err = this.addReconciler(required[0], this.Owning().GroupKind(), DEFAULT_POOL, DEFAULT_RECONCILER)
	if err != nil {
		return nil, err
	}

	for _, cmds := range this.GetDefinition().Commands() {
		for _, cmd := range cmds {
			_, err := this.addReconciler("", cmd.Key(), cmd.PoolName(), cmd.Reconciler())
			if err != nil {
				return nil, fmt.Errorf("Add matcher for reconciler %s failed: %s", cmd.Reconciler(), err)
			}
		}
	}
	for _, s := range this.GetDefinition().Syncers() {
		cluster := clusters.GetCluster(s.GetCluster())
		reconcilers := this.mappings.Get(cluster, s.GetResource().GroupKind())
		if len(reconcilers) == 0 {
			return nil, fmt.Errorf("resource %q not watched for cluster %s", s.GetResource(), s.GetCluster())
		}
		if err := this.syncRequests.AddSyncer(NewSyncer(s.GetName(), s.GetResource(), cluster)); err != nil {
			return nil, err
		}
		this.Infof("adding syncer %s for resource %s on cluster %s", s.GetName(), s.GetResource(), cluster)
	}

	return this, nil
}

func (this *controller) deployImplicitCustomResourceDefinitions(log logger.LogContext, eff WatchedResources, gks resources.GroupKindSet, cl cluster.Interface) error {
	for gk := range gks {
		if eff.Contains(cl.GetId(), gk) {
			eff.Remove(cl.GetId(), gk)
			v, err := apiextensions.NewDefaultedCustomResourceDefinitionVersions(gk)
			if err == nil {
				err := conditionalDeploy(v, log, cl, this.env.ControllerManager().GetMaintainer())
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (this *controller) deployCRDS() error {
	// first gather all intended or required resources
	// by its effective and logical cluster usage
	clusterResources := WatchedResources{}
	effClusterResources := WatchedResources{}
	if this.Owning() != nil {
		clusterResources.Add(CLUSTER_MAIN, this.Owning().GroupKind())
		effClusterResources.Add(this.GetMainCluster().GetId(), this.Owning().GroupKind())
	}
	for cname, watches := range this.definition.Watches() {
		cluster := this.GetCluster(cname)
		if cluster == nil {
			return fmt.Errorf("cluster %q not found for resource definitions", cname)
		}
		wctx := newWatchContext(this, cluster)
		for _, w := range watches {
			def := w.WatchResourceDef(wctx)
			key := def.Key
			if key != nil {
				clusterResources.Add(cname, key.GroupKind())
				effClusterResources.Add(cluster.GetId(), key.GroupKind())
			}
		}
	}
	for n, crds := range this.definition.CustomResourceDefinitions() {
		cluster := this.GetCluster(n)
		if cluster == nil {
			return fmt.Errorf("cluster %q not found for resource definitions", n)
		}
		for _, v := range crds {
			effClusterResources.Add(cluster.GetId(), v.GroupKind())
		}
	}

	// now deploy explicit requested CRDs or implicitly available CRDs for used resources
	log := this.AddIndent("  ")
	for n, crds := range this.definition.CustomResourceDefinitions() {
		cl := this.GetCluster(n)
		disabled, err := isDeployCRDsDisabled(cl)
		if err != nil {
			return err
		}
		if disabled {
			this.Infof("deployment of required crds is disabled for cluster %q (used for %q)", cl.GetName(), n)
			continue
		}
		this.Infof("ensure required crds for cluster %q (used for %q)", cl.GetName(), n)
		for _, v := range crds {
			clusterResources.Remove(n, v.GroupKind())
			if effClusterResources.Contains(cl.GetId(), v.GroupKind()) {
				effClusterResources.Remove(cl.GetId(), v.GroupKind())
				err := conditionalDeploy(v, log, cl, this.env.ControllerManager().GetMaintainer())
				if err != nil {
					return err
				}
			} else {
				log.Infof("crd for %s already handled", v.GroupKind())
			}
		}
		if err := this.deployImplicitCustomResourceDefinitions(log, effClusterResources, clusterResources[n], cl); err != nil {
			return err
		}
		delete(clusterResources, cl.GetId())
	}
	for n, gks := range clusterResources {
		cl := this.GetCluster(n)
		if len(gks) == 0 {
			continue
		}
		disabled, err := isDeployCRDsDisabled(cl)
		if err != nil {
			return err
		}
		if disabled {
			continue
		}
		this.Infof("ensure required crds for cluster %q (used for %q)", cl.GetName(), n)
		if err := this.deployImplicitCustomResourceDefinitions(log, effClusterResources, gks, cl); err != nil {
			return err
		}
	}
	return nil
}

func isDeployCRDsDisabled(cl cluster.Interface) (bool, error) {
	if cl.GetAttr(cluster.SUBOPTION_CONDITIONAL_DEPLOY_CRDS) == true {
		if cl.GetAttr(cluster.ConditionalDeployCRDIgnoreSetAttrKey) == nil {
			set, err := findCRDsDeployedByManagedResources(cl)
			if err != nil {
				return false, err
			}
			cl.SetAttr(cluster.ConditionalDeployCRDIgnoreSetAttrKey, set)
		}
		return false, nil
	}

	return cl.GetAttr(cluster.SUBOPTION_DISABLE_DEPLOY_CRDS) == true, nil
}

func (this *controller) whenReady() {
	this.ready.WhenReady()
}

func (this *controller) IsReady() bool {
	return this.ready.IsReady()
}

func (this *controller) GetReconciler(name string) reconcile.Interface {
	return this.reconcilers[name]
}

func (this *controller) addReconciler(cname string, spec ReconcilationElementSpec, pool string, reconciler string) (bool, error) {
	r := this.reconcilers[reconciler]
	if r == nil {
		return false, fmt.Errorf("reconciler %q not found for %q", reconciler, spec)
	}

	cluster := this.cluster
	cluster_name := ""
	aliases := utils.StringSet{}
	if cname != "" {
		cluster = this.clusters.GetCluster(cname)
		if cluster == nil {
			return false, fmt.Errorf("cluster %q not found for %q", mappings.ClusterName(cname), spec)
		}
		cluster_name = cluster.GetName()
		aliases = this.clusters.GetAliases(cluster.GetName())
	}
	if gk, ok := spec.(schema.GroupKind); ok {
		if reject, ok := r.(reconcile.ReconcilationRejection); ok {
			if reject.RejectResourceReconcilation(cluster, gk) {
				this.Infof("reconciler %s rejects resource reconcilation resource %s for cluster %s",
					reconciler, gk, cluster.GetName())
				return false, nil
			}
			this.Infof("reconciler %s supports reconcilation rejection and accepts resource %s for cluster %s",
				reconciler, gk, cluster.GetName())
		}
	}

	src := _ReconcilationKey{key: spec, cluster: cluster_name, reconciler: reconciler}
	mapping, ok := this.mappings[src]
	if ok {
		if mapping != pool {
			return false, fmt.Errorf("a key (%s) for the same cluster %q (used for %s) and reconciler (%s) can only be handled by one pool (found %q and %q)", spec, cluster_name, aliases, reconciler, pool, mapping)
		}
	} else {
		this.mappings[src] = pool
	}

	if cname == "" {
		this.Infof("*** adding reconciler %q for %q using pool %q", reconciler, spec, pool)
	} else {
		this.Infof("*** adding reconciler %q for %q in cluster %q (used for %q) using pool %q", reconciler, spec, cluster_name, mappings.ClusterName(cname), pool)
	}
	this.getPool(pool).addReconciler(spec, r)
	return true, nil
}

func (this *controller) getPool(name string) *pool {
	pool := this.pools[name]
	if pool == nil {
		def := this.definition.Pools()[name]

		if def == nil {
			panic(fmt.Sprintf("unknown pool %q for controller %q", name, this.GetName()))
		}
		this.Infof("get pool config %q", def.GetName())
		options := this.options.GetSource(def.GetName()).(config.OptionSet)
		size := options.GetOption(POOL_SIZE_OPTION).IntValue()
		period := def.Period()
		if period != 0 {
			period = options.GetOption(POOL_RESYNC_PERIOD_OPTION).DurationValue()
		}
		pool = NewPool(this, name, size, period)
		this.pools[name] = pool
	}
	return pool
}

func (this *controller) GetPool(name string) Pool {
	pool := this.pools[name]
	if pool == nil {
		return nil
	}
	return pool
}

func (this *controller) GetEnvironment() Environment {
	return this.env
}

func (this *controller) GetDefinition() Definition {
	return this.definition
}

func (this *controller) getClusterHandler(name string) (*ClusterHandler, error) {
	cluster := this.GetCluster(name)

	if cluster == nil {
		return nil, fmt.Errorf("unknown cluster %q for %q", name, this.GetName())
	}
	h := this.handlers[cluster.GetId()]
	if h == nil {
		var err error
		h, err = newClusterHandler(this, cluster)
		if err != nil {
			return nil, err
		}
		this.handlers[cluster.GetId()] = h
	}
	return h, nil
}

func (this *controller) ClusterHandler(cluster resources.Cluster) *ClusterHandler {
	return this.handlers[cluster.GetId()]
}

func (this *controller) GetClusterById(id string) cluster.Interface {
	return this.clusters.GetById(id)
}

func (this *controller) GetCluster(name string) cluster.Interface {
	if name == CLUSTER_MAIN || name == "" {
		return this.GetMainCluster()
	}
	return this.clusters.GetCluster(name)
}
func (this *controller) GetMainCluster() cluster.Interface {
	return this.cluster
}
func (this *controller) GetClusterAliases(eff string) utils.StringSet {
	return this.clusters.GetAliases(eff)
}
func (this *controller) GetEffectiveCluster(eff string) cluster.Interface {
	return this.clusters.GetEffective(eff)
}

func (this *controller) GetObject(key resources.ClusterObjectKey) (resources.Object, error) {
	return this.clusters.GetObject(key)
}

func (this *controller) GetCachedObject(key resources.ClusterObjectKey) (resources.Object, error) {
	return this.clusters.GetCachedObject(key)
}

func (this *controller) EnqueueKey(key resources.ClusterObjectKey) error {
	cluster := this.GetClusterById(key.Cluster())
	if cluster == nil {
		return fmt.Errorf("cluster with id %q not found", key.Cluster())
	}
	h := this.ClusterHandler(cluster)
	return h.EnqueueKey(key)
}

func (this *controller) Enqueue(object resources.Object) error {
	h := this.ClusterHandler(object.GetCluster())
	return h.EnqueueObject(object)
}

func (this *controller) EnqueueAfter(object resources.Object, duration time.Duration) error {
	h := this.ClusterHandler(object.GetCluster())
	return h.EnqueueObjectAfter(object, duration)
}

func (this *controller) EnqueueRateLimited(object resources.Object) error {
	h := this.ClusterHandler(object.GetCluster())
	return h.EnqueueObjectRateLimited(object)
}

func (this *controller) EnqueueCommand(cmd string) error {
	found := false
	for _, p := range this.pools {
		r := p.getReconcilers(cmd)
		if len(r) > 0 {
			p.EnqueueCommand(cmd)
			found = true
		}
	}
	if !found {
		return fmt.Errorf("no handler found for command %q", cmd)
	}
	return nil
}

func (this *controller) Owning() ResourceKey {
	return this.mainresc.Key
}

func (this *controller) GetMainWatchResource() WatchResource {
	return this.owning
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// controller start up

// Check does all the checks that might cause Prepare to fail
// after a successful check Prepare can execute without error
func (this *controller) check() error {
	h, err := this.getClusterHandler(CLUSTER_MAIN)
	if err != nil {
		return err
	}

	_, err = h.GetResource(this.Owning())
	if err != nil {
		return err
	}

	// setup and check cluster handlers for all required cluster
	for cname, watches := range this.watches {
		h, err := this.getClusterHandler(cname)
		if err != nil {
			return err
		}
		for _, watch := range watches {
			_, err = h.GetResource(watch.Key)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (this *controller) AddCluster(_ cluster.Interface) error {
	return nil
}

func (this *controller) registerWatch(h *ClusterHandler, r *watchDef) error {
	var optionsFunc resources.TweakListOptionsFunc
	if len(r.Tweaker) > 0 {
		optionsFunc = r.TweakListOptions
	}
	return h.register(r, r.Namespace, optionsFunc, this.getPool(r.PoolName))
}

func (this *controller) setup() error {
	this.Infof("setup reconcilers...")
	for n, r := range this.reconcilers {
		err := reconcile.SetupReconciler(r)
		if err != nil {
			return fmt.Errorf("setup of reconciler %s of controller %s failed: %s", n, this.GetName(), err)
		}
	}
	return nil
}

// Prepare finally prepares the controller to run
// all error conditions MUST also be checked
// in Check, so after a successful checkController
// startController MUST not return an error.
func (this *controller) prepare() error {
	h, err := this.getClusterHandler(CLUSTER_MAIN)
	if err != nil {
		return err
	}

	this.Infof("setup watches....")
	this.Infof("watching main resources %q at cluster %q (reconciler %s)", this.Owning(), h, DEFAULT_RECONCILER)

	if this.mainresc.Key != nil {
		err = this.registerWatch(h, this.mainresc)
		if err != nil {
			return err
		}
	}
	for cname, watches := range this.watches {
		h, err := this.getClusterHandler(cname)
		if err != nil {
			return err
		}

		for _, watch := range watches {
			this.Infof("watching additional resources %q at cluster %q (reconciler %s)", watch.Key, h, watch.Reconciler)
			err = this.registerWatch(h, watch)
			if err != nil {
				return err
			}
		}
	}
	this.Infof("setup watches done")

	return nil
}

func (this *controller) Run() {
	this.ready.ready()
	this.Infof("starting pools...")
	for _, p := range this.pools {
		ctxutil.WaitGroupRunAndCancelOnExit(this.GetContext(), p.Run)
	}

	this.Infof("starting reconcilers...")
	for n, r := range this.reconcilers {
		err := reconcile.StartReconciler(r)
		if err != nil {
			this.Errorf("exit controller %s because start of reconciler %s failed: %s", this.GetName(), n, err)
			return
		}
	}
	this.Infof("controller started")
	<-this.GetContext().Done()
	this.Info("waiting for worker pools to shutdown")
	ctxutil.WaitGroupWait(this.GetContext(), 120*time.Second)
	for n, r := range this.reconcilers {
		// TODO error handling
		_ = reconcile.CleanupReconciler(this, n, r)
	}
	this.Info("exit controller")
}

func (this *controller) mustHandle(r resources.Object) bool {
	for _, f := range this.filters {
		if !f(this.mainresc.Key, r) {
			this.Debugf("%s rejected by filter", r.Description())
			return false
		}
	}
	return true
}

func (this *controller) DecodeKey(key string) (string, *resources.ClusterObjectKey, resources.Object, error) {
	i := strings.Index(key, ":")

	if i < 0 {
		return key, nil, nil, nil
	}

	main := key[:i]
	if main == "cmd" {
		return key[i+1:], nil, nil, nil
	}
	if main == "obj" {
		key = key[i+1:]
	}
	i = strings.Index(key, ":")

	cluster := this.clusters.GetEffective(key[0:i])
	if cluster == nil {
		return "", nil, nil, fmt.Errorf("unknown cluster in key %q", key)
	}

	key = key[i+1:]

	apiGroup, kind, namespace, name, err := DecodeObjectSubKey(key)
	if err != nil {
		return "", nil, nil, fmt.Errorf("error decoding '%s': %s", key, err)
	}
	objKey := resources.NewClusterKey(cluster.GetId(), resources.NewGroupKind(apiGroup, kind), namespace, name)

	r, err := this.ClusterHandler(cluster).GetObject(objKey)
	return "", &objKey, r, err
}

func (this *controller) Synchronize(log logger.LogContext, name string, initiator resources.Object) (bool, error) {
	return this.syncRequests.Synchronize(log, name, initiator)
}

func (this *controller) requestHandled(log logger.LogContext, reconciler reconcile.Interface, key resources.ClusterObjectKey) {
	this.syncRequests.requestHandled(log, this.reconcilerNames[reconciler], key)
}

func (this *controller) leaseMeta(name string, cnames ...string) (leasename string, cname string, c cluster.Interface) {
	cname = CLUSTER_MAIN
	if len(cnames) > 0 {
		cname = cnames[0]
	}
	leasename = this.env.GetConfig().Lease.LeaseName + "-" + this.GetName() + "-" + name
	c = this.GetCluster(cname)
	return
}

func (this *controller) HasLeaseRequest(name string, cnames ...string) bool {
	leasename, _, c := this.leaseMeta(name, cnames...)
	if c == nil {
		return false
	}

	this.lock.Lock()
	defer this.lock.Unlock()
	cl := this.leases[c.GetId()]
	if cl == nil {
		return false
	}
	_, ok := cl[leasename]
	return ok
}

func (this *controller) IsLeaseActive(name string, cnames ...string) bool {
	leasename, _, c := this.leaseMeta(name, cnames...)
	if c == nil {
		return false
	}

	this.lock.Lock()
	defer this.lock.Unlock()
	cl := this.leases[c.GetId()]
	return cl != nil && cl[leasename]
}

func (this *controller) WithLease(name string, regain bool, action func(ctx context.Context), cnames ...string) error {
	leasename, cname, c := this.leaseMeta(name, cnames...)
	if c == nil {
		return fmt.Errorf("unknown cluster %q", cname)
	}

	cfg := this.env.GetConfig().Lease
	cfg.LeaseName = leasename

	leaderElectionConfig, err := lease.MakeLeaderElectionConfig(c, this.env.Namespace(), &cfg)
	if err != nil {
		return err
	}
	this.lock.Lock()
	defer this.lock.Unlock()

	cl := this.leases[c.GetId()]
	if cl == nil {
		cl = map[string]bool{}
		this.leases[c.GetId()] = cl
	}
	_, ok := cl[cfg.LeaseName]
	if ok {
		return fmt.Errorf("lease request already pending")
	}

	leasectx := ctxutil.CancelContext(this.GetContext())
	leaderElectionConfig.Callbacks = leaderelection.LeaderCallbacks{
		OnStartedLeading: func(ctx context.Context) {
			this.lock.Lock()
			cl[cfg.LeaseName] = true
			this.lock.Unlock()

			this.Infof("starting lease function for %s", cfg.LeaseName)
			action(ctx)
			this.lock.Lock()
			defer this.lock.Unlock()
			delete(cl, cfg.LeaseName)
			select {
			case <-ctx.Done():
				this.Infof("lease function for %s stopped", cfg.LeaseName)
			default:
				this.Infof("lease function for %s done -> stopping lease", cfg.LeaseName)
				ctxutil.Cancel(leasectx)
			}
		},
		OnStoppedLeading: func() {
			this.Infof("lost leadership %s.", cfg.LeaseName)
			this.lock.Lock()
			defer this.lock.Unlock()
			if _, ok := cl[cfg.LeaseName]; ok {
				cl[cfg.LeaseName] = false
			}
		},
	}

	leaderElector, err := leaderelection.NewLeaderElector(*leaderElectionConfig)
	if err != nil {
		return fmt.Errorf("couldn't create leader elector: %v", err)
	}

	cl[cfg.LeaseName] = false
	go func() {
		ctxutil.WaitGroupAdd(leasectx)
		defer ctxutil.WaitGroupDone(leasectx)
		for {
			this.Infof("requesting lease execution %q for cluster %s in namespace %q",
				cfg.LeaseName, c, this.env.Namespace())

			leaderElector.Run(leasectx)
			this.Infof("lease %s gone", cfg.LeaseName)
			select {
			case <-leasectx.Done():
				return
			default:
				if !regain {
					this.Infof("stopping controller manager because of lost lease %s", cfg.LeaseName)
					ctxutil.Cancel(this.env.ControllerManager().GetContext())
					return
				}
			}
		}
	}()
	return nil
}

func conditionalDeploy(v *apiextensions.CustomResourceDefinitionVersions, log logger.LogContext, cl cluster.Interface, maintainerInfo extension.MaintainerInfo) error {
	ignoreSet := set.Set[string]{}
	if v := cl.GetAttr(cluster.ConditionalDeployCRDIgnoreSetAttrKey); v != nil {
		maintainerInfo.ForceCRDUpdate = true
		ignoreSet = v.(set.Set[string])
	}

	if ignoreSet.Has(v.Name()) {
		log.Infof("ignoring CRD managed by Gardener managed resources: %s", v.GroupKind())
		return nil
	}

	return v.Deploy(log, cl, maintainerInfo)
}

func findCRDsDeployedByManagedResources(cl cluster.Interface) (set.Set[string], error) {
	resns, err := cl.Resources().GetByExample(&corev1.Namespace{})
	if err != nil {
		return nil, fmt.Errorf("findCRDsDeployedByManagedResources: could not get resources for namespaces: %w", err)
	}

	ns := &corev1.Namespace{}
	if _, err := resns.GetInto(resources.NewObjectName("garden"), ns); err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("findCRDsDeployedByManagedResources: could not get namespace: %w", err)
	}

	resmr, err := cl.Resources().GetByExample(&resourcesv1alpha1.ManagedResource{})
	if err != nil {
		if resourceserrors.IsKind(resourceserrors.ERR_UNKNOWN_RESOURCE, err) {
			return nil, nil
		}
		return nil, fmt.Errorf("findCRDsDeployedByManagedResources: could not get resources for unstructured managed resource: %w", err)
	}

	objs, err := resmr.Namespace("garden").List(metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("findCRDsDeployedByManagedResources: could not list managed resources: %w", err)
	}
	crdNames := set.Set[string]{}
	for _, obj := range objs {
		mr, ok := obj.Data().(*resourcesv1alpha1.ManagedResource)
		if !ok {
			return nil, fmt.Errorf("findCRDsDeployedByManagedResources: could not cast object to ManagedResource: %t", obj.Data())
		}
		for _, item := range mr.Status.Resources {
			if item.APIVersion != "apiextensions.k8s.io/v1" {
				continue
			}
			if item.Kind != "CustomResourceDefinition" {
				continue
			}
			crdNames.Insert(item.Name)
		}
	}
	return crdNames, nil
}
