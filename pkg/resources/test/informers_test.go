/*
 * SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package test_test

import (
	"context"
	"reflect"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	kcorev1 "k8s.io/api/core/v1"
	kmetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gardener/controller-manager-library/pkg/resources"
)

type Object interface {
	kmetav1.Object
	runtime.Object
}

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
	cl, err := client.New(cfg, client.Options{})
	Expect(err).NotTo(HaveOccurred())
	err = cl.Create(context.Background(), secret)
	Expect(err).NotTo(HaveOccurred())
	return secret
}

func deleteSecret(secret Object) {
	if secret == nil {
		return
	}
	cl, err := client.New(cfg, client.Options{})
	Expect(err).NotTo(HaveOccurred())
	err = cl.Delete(context.Background(), secret)
	Expect(err).NotTo(HaveOccurred())
}

var _ = Describe("Informers", func() {
	var (
		knownSecret1 Object
		knownSecret2 Object

		gvk = schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Secret"}
	)

	BeforeEach(func() {
		Expect(cfg).NotTo(BeNil())
		knownSecret1 = createSecret("informers-test-1", "default")

		info, err := defaultCluster.ResourceContext().Get(gvk)
		Expect(err).NotTo(HaveOccurred())
		Expect(info).NotTo(BeNil())
	})

	AfterEach(func() {
		By("cleaning up created secrets")
		deleteSecret(knownSecret1)
		deleteSecret(knownSecret2)
	})

	Describe("informer list and watch test", func() {
		var err error
		var rStructured, rUnstructured resources.Interface
		var foundWatchForSecret2 bool

		It("should handle list and watch for all supported object types", func() {
			By("list structured objects", func() {
				rStructured, err = defaultCluster.Resources().GetByGVK(gvk)
				Expect(err).NotTo(HaveOccurred())
				Expect(rStructured).NotTo(BeNil())

				list, err := rStructured.ListCached(nil)
				Expect(err).NotTo(HaveOccurred())
				checkSecret(list, knownSecret1, reflect.TypeOf(&kcorev1.Secret{}))
			})
			By("list unstructured objects", func() {
				rUnstructured, err = defaultCluster.Resources().GetUnstructuredByGVK(gvk)
				Expect(err).NotTo(HaveOccurred())
				Expect(rUnstructured).NotTo(BeNil())

				list, err := rUnstructured.ListCached(nil)
				Expect(err).NotTo(HaveOccurred())
				checkSecret(list, knownSecret1, reflect.TypeOf(&unstructured.Unstructured{}))
			})
			By("add secret and wait for watch", func() {
				rStructured.AddInfoEventHandler(resources.ResourceInfoEventHandlerFuncs{
					AddFunc: func(obj resources.ObjectInfo) {
						if !foundWatchForSecret2 {
							foundWatchForSecret2 = obj.Key().Name() == "informers-test-2"
						}
					},
				})
				knownSecret2 = createSecret("informers-test-2", "default")
			})
			By("watch with event handler", func() {
				func() {
					for i := 0; i < 50; i++ {
						if foundWatchForSecret2 {
							return
						}
						time.Sleep(50 * time.Millisecond)
					}
				}()
				Expect(foundWatchForSecret2).To(BeTrue())
			})
			By("watch structured objects", func() {
				waitForSecret(rStructured, knownSecret2.GetName())
				list, err := rStructured.ListCached(nil)
				Expect(err).NotTo(HaveOccurred())
				checkSecret(list, knownSecret1, reflect.TypeOf(&kcorev1.Secret{}))
				checkSecret(list, knownSecret2, reflect.TypeOf(&kcorev1.Secret{}))
			})
			By("watch unstructured objects", func() {
				waitForSecret(rUnstructured, knownSecret2.GetName())
				list, err := rUnstructured.ListCached(nil)
				Expect(err).NotTo(HaveOccurred())
				checkSecret(list, knownSecret1, reflect.TypeOf(&unstructured.Unstructured{}))
				checkSecret(list, knownSecret2, reflect.TypeOf(&unstructured.Unstructured{}))
			})
		})
	})
})

func waitForSecret(res resources.Interface, name string) {
	selector, err := kmetav1.LabelSelectorAsSelector(&kmetav1.LabelSelector{
		MatchLabels: map[string]string{"test-label": name},
	})
	Expect(err).NotTo(HaveOccurred())
	for i := 0; i < 50; i++ {
		list, err := res.ListCached(selector)
		Expect(err).NotTo(HaveOccurred())
		if len(list) > 0 {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	Fail("secret not found")
}

func checkSecret(list []resources.Object, expectedSecret Object, expectedType reflect.Type) {
	Expect(list).NotTo(BeEmpty())
	found := false
	for _, obj := range list {
		actualType := reflect.TypeOf(obj.Data())
		Expect(actualType).To(Equal(expectedType))
		if obj.GetName() == expectedSecret.GetName() && obj.GetNamespace() == expectedSecret.GetNamespace() {
			found = true
		}
	}
	Expect(found).To(BeTrue(), "not found: "+expectedSecret.GetName())
}
