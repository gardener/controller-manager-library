/*
 * SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package test_test

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/gardener/controller-manager-library/pkg/resources"
)

func TestSource(t *testing.T) {
	RegisterFailHandler(Fail)
	reporterConfig := types.NewDefaultReporterConfig()
	reporterConfig.SlowSpecThreshold = 1 * time.Minute
	RunSpecs(t, "ControllerManager Suite", reporterConfig)
}

var (
	testenv        *envtest.Environment
	restConfig     *rest.Config
	clientset      *kubernetes.Clientset
	kubeconfigFile string
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

	testenv = &envtest.Environment{}

	var err error
	restConfig, err = testenv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(restConfig).ToNot(BeNil())
	kubeconfigFile = createKubeconfigFile(restConfig)
	println("KUBECONFIG=" + kubeconfigFile)

	resources.Register(corev1.SchemeBuilder)

	clientset, err = kubernetes.NewForConfig(restConfig)
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	if len(kubeconfigFile) > 0 {

	}
	_ = os.Remove(kubeconfigFile)
	By("stopping test environment")
	Expect(testenv.Stop()).To(Succeed())
})

func createKubeconfigFile(cfg *rest.Config) string {
	template := `apiVersion: v1
kind: Config
clusters:
  - name: testenv
    cluster:
      server: '%s'
      certificate-authority-data: %s
contexts:
  - name: testenv
    context:
      cluster: testenv
      user: testuser
current-context: testenv
users:
  - name: testuser
    user:
      client-certificate-data: %s
      client-key-data: %s`

	tmpfile, err := ioutil.TempFile("", "kubeconfig-controllermanager-suite-test")
	Expect(err).NotTo(HaveOccurred())
	_, err = fmt.Fprintf(tmpfile, template, cfg.Host, base64.StdEncoding.EncodeToString(cfg.CAData),
		base64.StdEncoding.EncodeToString(cfg.CertData), base64.StdEncoding.EncodeToString(cfg.KeyData))
	Expect(err).NotTo(HaveOccurred())
	err = tmpfile.Close()
	Expect(err).NotTo(HaveOccurred())
	return tmpfile.Name()
}
