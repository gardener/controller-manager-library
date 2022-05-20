/*
 * SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package test_test

import (
	"context"
	"testing"

	"github.com/gardener/controller-manager-library/pkg/logger"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/gardener/controller-manager-library/pkg/controllermanager/cluster"
	"github.com/gardener/controller-manager-library/pkg/resources"
)

func TestSource(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Resources Suite")
}

var (
	testenv           *envtest.Environment
	restConfig        *rest.Config
	clientset         *kubernetes.Clientset
	defaultCluster    cluster.Interface
	defaultClusterCtx context.Context
	cancel            context.CancelFunc
	logr              *logrus.Entry
)

func init() {
	// allows to set -v=8 for REST request logging
	klog.InitFlags(nil)
}

var _ = BeforeSuite(func() {
	// enable manager logs
	logf.SetLogger(zap.New(zap.UseDevMode(true), zap.WriteTo(GinkgoWriter)))
	log := logrus.New()
	log.SetOutput(GinkgoWriter)
	logr = logrus.NewEntry(log)

	testenv = &envtest.Environment{}
	var err error
	restConfig, err = testenv.Start()
	Expect(err).NotTo(HaveOccurred())

	resources.Register(corev1.SchemeBuilder)
	def := cluster.Configure("default", "kubeconfig", "dummy").Definition()
	defaultClusterCtx, cancel = context.WithCancel(context.Background())
	defaultCluster, err = cluster.CreateClusterForScheme(defaultClusterCtx, logger.New(), def, "default", restConfig, nil)
	Expect(err).NotTo(HaveOccurred())

	clientset, err = kubernetes.NewForConfig(restConfig)
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	cancel()

	By("stopping test environment")
	Expect(testenv.Stop()).To(Succeed())
})
