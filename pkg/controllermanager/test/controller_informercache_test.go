/*
 * SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package test_test

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"time"

	. "github.com/onsi/ginkgo"
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
	cl, err := client.New(cfg, client.Options{})
	Expect(err).NotTo(HaveOccurred())
	err = cl.Create(context.Background(), obj)
	if ignoreAlreadyExists && err != nil && errors.IsAlreadyExists(err) {
		return
	}
	Expect(err).NotTo(HaveOccurred())
}

func updateObject(obj Object) {
	cl, err := client.New(cfg, client.Options{})
	Expect(err).NotTo(HaveOccurred())
	err = cl.Update(context.Background(), obj)
	Expect(err).NotTo(HaveOccurred())
}

func deleteObject(obj Object, ignoreNotFound bool) {
	if obj == nil {
		return
	}
	cl, err := client.New(cfg, client.Options{})
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
	secret1       Object
	secret2       Object
	testNamespace Object

	setup   bool
	started bool

	reconcileCountSecret1 int
	reconcileCountSecret2 int
	deletedCountSecret2   int
	lastError             error
}

var _ = Describe("Informers", func() {
	var (
		ctxCM context.Context

		data *reconcilerData
	)

	BeforeEach(func() {
		Expect(cfg).NotTo(BeNil())
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
		deleteObject(data.secret2, true)
	})

	var startControllerManager = func(testkind MinimalWatchTestKind) {
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
		timeout := (3 + bigSecretsCount/2) * time.Second
		startControllerManager(testkind)
		waitFor("setup of controller", func() bool { return data.setup }, 1*time.Second)
		waitFor("start of controller", func() bool { return data.started }, 1*time.Second)

		waitFor("reconciling existing secret1", func() bool { return data.reconcileCountSecret1 > 0 }, timeout)

		By("create secret2", func() {
			minimal.ConvertCounter = 0
			data.testNamespace = createNamespace("test")
			data.secret2 = createSecret("informers-test-2", "test")
		})
		waitFor("reconciling new secret2", func() bool { return data.reconcileCountSecret2 > 0 }, timeout)

		oldCount := data.reconcileCountSecret2
		By("update secret2", func() {
			s := data.secret2.(*kcorev1.Secret)
			s.Data["foo2"] = []byte("blabla")
			updateObject(data.secret2)
		})
		waitFor("reconciling updated secret2", func() bool { return data.reconcileCountSecret2 > oldCount }, timeout)

		By("delete secret2", func() {
			deleteObject(data.secret2, false)
		})
		waitFor("watch deleted secret2", func() bool { return data.deletedCountSecret2 > 0 }, timeout)

		oldCount = data.reconcileCountSecret1
		waitFor("periodic reconciling secret1", func() bool { return data.reconcileCountSecret1 > oldCount }, 5*time.Second+timeout)

		Expect(data.lastError).NotTo(HaveOccurred())

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
	h.data.setup = true
	return nil
}

func (h *reconciler) Start() error {
	h.data.started = true
	return nil
}

func (h *reconciler) Command(logger logger.LogContext, cmd string) reconcile.Status {
	logger.Infof("got command %q", cmd)
	return reconcile.Succeeded(logger)
}

func (h *reconciler) Reconcile(logger logger.LogContext, obj resources.Object) reconcile.Status {
	logger.Infof("reconcile %s", obj.ObjectName())

	check := func(candidate Object, count *int) {
		if obj.GetName() == candidate.GetName() && obj.GetNamespace() == candidate.GetNamespace() {
			(*count)++
			if !reflect.DeepEqual(obj.Data().(*kcorev1.Secret).Data, candidate.(*kcorev1.Secret).Data) {
				h.data.lastError = fmt.Errorf("secret %s data mismatch", candidate.GetName())
			}
		}
	}
	check(h.data.secret1, &h.data.reconcileCountSecret1)
	if h.data.secret2 != nil {
		check(h.data.secret2, &h.data.reconcileCountSecret2)
	}

	return reconcile.Succeeded(logger)
}

func (h *reconciler) Delete(logger logger.LogContext, obj resources.Object) reconcile.Status {
	logger.Infof("delete %s", obj.ObjectName())
	return reconcile.Succeeded(logger)
}

func (h *reconciler) Deleted(logger logger.LogContext, key resources.ClusterObjectKey) reconcile.Status {
	logger.Infof("deleted %s", key.ObjectName())
	if h.data.secret2 != nil {
		if key.Name() == h.data.secret2.GetName() && key.Namespace() == h.data.secret2.GetNamespace() {
			h.data.deletedCountSecret2++
		}
	}
	return reconcile.Succeeded(logger)
}
