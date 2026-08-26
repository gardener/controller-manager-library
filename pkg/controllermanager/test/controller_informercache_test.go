/*
 * SPDX-FileCopyrightText: Copyright Contributors to the Gardener project
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package test_test

import (
	"context"
	"fmt"
	"maps"
	"reflect"
	"runtime"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kcorev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	kmetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gardener/controller-manager-library/pkg/configmain"
	"github.com/gardener/controller-manager-library/pkg/controllermanager"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/cluster"
	config2 "github.com/gardener/controller-manager-library/pkg/controllermanager/config"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/controller"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/controller/reconcile"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/controller/reconcile/reconcilers"
	"github.com/gardener/controller-manager-library/pkg/ctxutil"
	"github.com/gardener/controller-manager-library/pkg/logger"
	"github.com/gardener/controller-manager-library/pkg/resources"
	"github.com/gardener/controller-manager-library/pkg/resources/minimal"
)

type Object resources.ObjectData

type MinimalWatchTestKind int

const (
	Normal MinimalWatchTestKind = iota
	GloballyMinimalWatch
	ControllerMinimalWatch
)

// set to 100 and run only cached or uncached to see difference in memory consumption
const bigSecretsCount = 0

func createSecret(name, namespace string) Object {
	secret := &kcorev1.Secret{
		ObjectMeta: kmetav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"test-label": name,
			},
		},
		Data: map[string][]byte{
			"foo": []byte("bar-" + name),
		},
	}
	createObject(secret, false)
	return secret
}

func createBigSecret(n int) Object {
	fooData := make([]byte, 100000)
	for j := 0; j < len(fooData); j++ {
		fooData[j] = byte(65 + j%26)
	}
	secret := &kcorev1.Secret{
		ObjectMeta: kmetav1.ObjectMeta{
			Name:      fmt.Sprintf("big-%d", n),
			Namespace: "default",
			Labels: map[string]string{
				"test-label": "bigsecret",
			},
		},
		Data: map[string][]byte{
			"foo": fooData,
		},
	}
	createObject(secret, false)
	return secret
}

func deleteBigSecret(n int) {
	secret := &kcorev1.Secret{
		ObjectMeta: kmetav1.ObjectMeta{
			Name:      fmt.Sprintf("big-%d", n),
			Namespace: "default",
			Labels: map[string]string{
				"test-label": "bigsecret",
			},
		},
		Data: map[string][]byte{},
	}
	deleteObject(secret, true)
}

func createNamespace(namespace string) Object {
	ns := &kcorev1.Namespace{
		ObjectMeta: kmetav1.ObjectMeta{
			Name: namespace,
		},
	}
	createObject(ns, true)
	return ns
}

func createObject(obj Object, ignoreAlreadyExists bool) {
	cl, err := client.New(restConfig, client.Options{})
	Expect(err).NotTo(HaveOccurred())
	err = cl.Create(context.Background(), obj)
	if ignoreAlreadyExists && err != nil && errors.IsAlreadyExists(err) {
		return
	}
	Expect(err).NotTo(HaveOccurred())
}

// updateObject fetches a fresh copy of obj, applies mutate to it and updates
// it at the server. It returns the updated object (carrying the new resource
// version). The shared object passed in is never mutated, so it cannot briefly
// diverge from the server state and race a reconcile in flight.
func updateObject(obj Object, mutate func(obj Object)) Object {
	cl, err := client.New(restConfig, client.Options{})
	Expect(err).NotTo(HaveOccurred())
	ctx := context.Background()
	fresh := obj.DeepCopyObject().(Object)
	err = cl.Get(ctx, client.ObjectKeyFromObject(obj), fresh)
	Expect(err).NotTo(HaveOccurred())
	mutate(fresh)
	err = cl.Update(ctx, fresh)
	Expect(err).NotTo(HaveOccurred())
	return fresh
}

func deleteObject(obj Object, ignoreNotFound bool) {
	if obj == nil {
		return
	}
	cl, err := client.New(restConfig, client.Options{})
	Expect(err).NotTo(HaveOccurred())
	err = cl.Delete(context.Background(), obj)
	if ignoreNotFound && err != nil && errors.IsNotFound(err) {
		return
	}
	Expect(err).NotTo(HaveOccurred())
	dummy := obj.DeepCopyObject().(client.Object)
	for i := 0; i < 100; i++ {
		err = cl.Get(context.Background(), types.NamespacedName{Namespace: obj.GetNamespace(), Name: obj.GetName()}, dummy)
		time.Sleep(10 * time.Millisecond)
		if err != nil && errors.IsNotFound(err) {
			return
		}
	}
	Fail("deletion of " + obj.GetName() + " failed")
}

type reconcilerData struct {
	// lock guards all fields below; they are read from the test goroutine
	// and written from the controller's reconciler goroutine (and vice versa).
	lock sync.Mutex

	secret1       Object
	secret2       Object
	testNamespace Object
	testKind      MinimalWatchTestKind

	// expectedSecret2Data is the data the test last successfully wrote to
	// secret2 at the server, together with the resource version that write
	// produced. The reconciler only asserts on secret2 data when the object
	// it observes carries exactly this resource version.
	expectedSecret2Data    map[string][]byte
	expectedSecret2Version string

	setup   bool
	started bool

	reconcileCountSecret1 int
	reconcileCountSecret2 int
	deletedCountSecret2   int
	lastError             error
}

func (d *reconcilerData) getBool(p *bool) bool {
	d.lock.Lock()
	defer d.lock.Unlock()
	return *p
}

func (d *reconcilerData) getInt(p *int) int {
	d.lock.Lock()
	defer d.lock.Unlock()
	return *p
}

func (d *reconcilerData) getSecret2() Object {
	d.lock.Lock()
	defer d.lock.Unlock()
	return d.secret2
}

func (d *reconcilerData) getLastError() error {
	d.lock.Lock()
	defer d.lock.Unlock()
	return d.lastError
}

var _ = Describe("Informers", func() {
	var (
		ctxCM context.Context

		data *reconcilerData
	)

	BeforeEach(func() {
		Expect(restConfig).NotTo(BeNil())
		controller.ResetRegistryForTesting()

		data = &reconcilerData{}
		data.secret1 = createSecret("informers-test-1", "default")
		for i := 0; i < bigSecretsCount; i++ {
			createBigSecret(i)
		}
	})

	AfterEach(func() {
		if ctxCM != nil {
			ctxutil.Cancel(ctxCM)
		}
		By("cleaning up created secrets")
		for i := 0; i < bigSecretsCount; i++ {
			deleteBigSecret(i)
		}
		deleteObject(data.secret1, false)
		deleteObject(data.getSecret2(), true)
	})

	var startControllerManager = func(testkind MinimalWatchTestKind) {
		data.testKind = testkind
		createReconciler := func(c controller.Interface) (reconcile.Interface, error) {
			return &reconciler{
				controller: c,
				data:       data,
			}, nil
		}

		gvk := schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Secret"}

		ctrlconfig := controller.Configure("secrets").
			Reconciler(createReconciler).
			DefaultWorkerPool(1, 5*time.Second).
			MainResourceByGK(gvk.GroupKind()).
			With(reconcilers.SecretUsageReconciler(controller.CLUSTER_MAIN))

		ctx00 := ctxutil.CancelContext(ctxutil.WaitGroupContext(context.Background(), "main"))
		ctx0 := ctxutil.TickContext(ctx00, controllermanager.DeletionActivity)
		var cfg *configmain.Config
		ctxCM, cfg = configmain.WithConfig(ctx0, nil)
		Expect(cfg).NotTo(BeNil())
		minimalWatches := []schema.GroupKind{}
		switch testkind {
		case GloballyMinimalWatch:
			minimalWatches = []schema.GroupKind{gvk.GroupKind()}
		case ControllerMinimalWatch:
			ctrlconfig = ctrlconfig.MinimalWatches(gvk.GroupKind())
		}

		ctrlconfig.MustRegister()
		def := controllermanager.PrepareStart("informercache-test", "").GlobalMinimalWatch(minimalWatches...).Definition()
		def.ExtendConfig(cfg)
		df := cfg.GetSource("controllermanager").(*config2.Config)
		df2 := df.GetSource("cluster.default").(*cluster.Config)
		df2.KubeConfig = kubeconfigFile

		controllerManager, err := controllermanager.NewControllerManager(ctxCM, def)
		Expect(err).NotTo(HaveOccurred())
		go func() {
			defer GinkgoRecover()
			err := controllerManager.Run()
			Expect(err).NotTo(HaveOccurred())
		}()
	}

	var testFunc = func(testkind MinimalWatchTestKind) {
		timeout := (5 + bigSecretsCount/2) * time.Second
		startControllerManager(testkind)
		waitFor("setup of controller", func() bool { return data.getBool(&data.setup) }, 1*time.Second)
		waitFor("start of controller", func() bool { return data.getBool(&data.started) }, 1*time.Second)

		waitFor("reconciling existing secret1", func() bool { return data.getInt(&data.reconcileCountSecret1) > 0 }, timeout)

		By("create secret2", func() {
			minimal.ConvertCounter = 0
			data.testNamespace = createNamespace("test")
			secret2 := createSecret("informers-test-2", "test")
			data.lock.Lock()
			data.secret2 = secret2
			data.expectedSecret2Data = maps.Clone(secret2.(*kcorev1.Secret).Data)
			data.expectedSecret2Version = secret2.GetResourceVersion()
			data.lock.Unlock()
		})
		waitFor("reconciling new secret2", func() bool { return data.getInt(&data.reconcileCountSecret2) > 0 }, timeout)

		oldCount := data.getInt(&data.reconcileCountSecret2)
		By("update secret2", func() {
			updated := updateObject(data.getSecret2(), func(obj Object) {
				obj.(*kcorev1.Secret).Data["foo2"] = []byte("blabla")
			})
			// After the server update succeeded, publish the new expected
			// data/version so the reconciler asserts against exactly the
			// version it observes.
			data.lock.Lock()
			data.expectedSecret2Data = maps.Clone(updated.(*kcorev1.Secret).Data)
			data.expectedSecret2Version = updated.GetResourceVersion()
			data.lock.Unlock()
		})
		waitFor("reconciling updated secret2", func() bool { return data.getInt(&data.reconcileCountSecret2) > oldCount }, timeout)

		By("delete secret2", func() {
			deleteObject(data.getSecret2(), false)
		})
		waitFor("watch deleted secret2", func() bool { return data.getInt(&data.deletedCountSecret2) > 0 }, timeout)

		oldCount = data.getInt(&data.reconcileCountSecret1)
		waitFor("periodic reconciling secret1", func() bool { return data.getInt(&data.reconcileCountSecret1) > oldCount }, 5*time.Second+timeout)

		Expect(data.getLastError()).NotTo(HaveOccurred())

		switch testkind {
		case Normal:
			Expect(minimal.ConvertCounter).To(Equal(0))
		default:
			Expect(minimal.ConvertCounter).NotTo(Equal(0))
		}

		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		logger.Infof("memory: %d kB", m.HeapAlloc/1024)
	}

	Describe("controller watch", func() {
		It("should work for full watches", func() {
			testFunc(Normal)
		})
		It("should work for minimal watches", func() {
			testFunc(GloballyMinimalWatch)
		})
		It("should work for minimal watches (defined on controller level)", func() {
			testFunc(ControllerMinimalWatch)
		})
	})
})

func waitFor(msg string, check func() bool, timeout time.Duration) {
	max := int(timeout / (10 * time.Millisecond))
	for i := 0; i < max; i++ {
		if check() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	Fail(msg)
}

type reconciler struct {
	reconcile.DefaultReconciler
	controller controller.Interface
	data       *reconcilerData
}

func (h *reconciler) Setup() error {
	h.data.lock.Lock()
	h.data.setup = true
	h.data.lock.Unlock()
	return nil
}

func (h *reconciler) Start() error {
	h.data.lock.Lock()
	h.data.started = true
	h.data.lock.Unlock()
	return nil
}

func (h *reconciler) Command(logger logger.LogContext, cmd string) reconcile.Status {
	logger.Infof("got command %q", cmd)
	return reconcile.Succeeded(logger)
}

func (h *reconciler) Reconcile(logger logger.LogContext, obj resources.Object) reconcile.Status {
	logger.Infof("reconcile %s", obj.ObjectName())

	h.data.lock.Lock()
	secret1 := h.data.secret1
	secret2 := h.data.secret2
	testKind := h.data.testKind
	expectedSecret2Data := maps.Clone(h.data.expectedSecret2Data)
	expectedSecret2Version := h.data.expectedSecret2Version
	h.data.lock.Unlock()

	setErr := func(err error) {
		h.data.lock.Lock()
		if h.data.lastError == nil {
			h.data.lastError = err
		}
		h.data.lock.Unlock()
	}
	incCount := func(p *int) {
		h.data.lock.Lock()
		*p++
		h.data.lock.Unlock()
	}

	// check verifies an observed object against the expected data. expectedData
	// is the data the object should carry; expectedVersion, when non-empty,
	// gates the data assertion: it is only performed when the observed object's
	// resource version matches expectedVersion. This avoids false mismatches
	// when the reconciler observes a version the test has already superseded
	// (redelivery, resync, or a reconcile triggered by another reconciler).
	check := func(candidate Object, expectedData map[string][]byte, expectedVersion string, count *int) {
		if candidate == nil ||
			obj.GetName() != candidate.GetName() || obj.GetNamespace() != candidate.GetNamespace() {
			return
		}
		incCount(count)
		switch testKind {
		case Normal:
			if obj.IsMinimal() {
				setErr(fmt.Errorf("secret object %s unexpected minimal", candidate.GetName()))
				return
			}
			if expectedVersion != "" && obj.GetResourceVersion() != expectedVersion {
				return
			}
			if !reflect.DeepEqual(obj.Data().(*kcorev1.Secret).Data, expectedData) {
				setErr(fmt.Errorf("secret %s data mismatch", candidate.GetName()))
			}
		case ControllerMinimalWatch, GloballyMinimalWatch:
			if !obj.IsMinimal() {
				setErr(fmt.Errorf("secret object %s unexpected not minimal", candidate.GetName()))
				return
			}
			if obj.MinimalData() == nil {
				setErr(fmt.Errorf("secret %s unexpected data type: %T", obj.Data(), candidate.GetName()))
				return
			}
			realObj, err := obj.GetFullObject()
			if err != nil {
				setErr(err)
				return
			}
			// GetFullObject does a live API GET returning the latest server
			// state, which is not necessarily the version this event was
			// delivered for. Only assert data when the minimal event, the full
			// object, and the expected version all agree; otherwise this
			// reconcile is racing a newer write and the assertion is moot.
			if expectedVersion != "" &&
				(obj.MinimalData().GetResourceVersion() != expectedVersion ||
					realObj.GetResourceVersion() != expectedVersion) {
				return
			}
			if !reflect.DeepEqual(realObj.Data().(*kcorev1.Secret).Data, expectedData) {
				setErr(fmt.Errorf("secret %s data mismatch", candidate.GetName()))
			}
		}
	}

	// secret1 is never modified after creation, so its data is stable and we do
	// not gate on a version.
	if secret1 != nil {
		check(secret1, secret1.(*kcorev1.Secret).Data, "", &h.data.reconcileCountSecret1)
	}
	if secret2 != nil {
		check(secret2, expectedSecret2Data, expectedSecret2Version, &h.data.reconcileCountSecret2)
	}

	return reconcile.Succeeded(logger)
}

func (h *reconciler) Delete(logger logger.LogContext, obj resources.Object) reconcile.Status {
	logger.Infof("delete %s", obj.ObjectName())
	return reconcile.Succeeded(logger)
}

func (h *reconciler) Deleted(logger logger.LogContext, key resources.ClusterObjectKey) reconcile.Status {
	logger.Infof("deleted %s", key.ObjectName())
	h.data.lock.Lock()
	defer h.data.lock.Unlock()
	if h.data.secret2 != nil {
		if key.Name() == h.data.secret2.GetName() && key.Namespace() == h.data.secret2.GetNamespace() {
			h.data.deletedCountSecret2++
		}
	}
	return reconcile.Succeeded(logger)
}
