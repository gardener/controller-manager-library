module github.com/gardener/controller-manager-library

go 1.16

require (
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Masterminds/semver/v3 v3.1.1
	github.com/Masterminds/sprig v2.22.0+incompatible // indirect
	github.com/ahmetb/gen-crd-api-reference-docs v0.2.0
	github.com/appscode/jsonpatch v1.0.1
	github.com/cyphar/filepath-securejoin v0.2.2 // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/huandu/xstrings v1.2.0 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.15.0
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.2.1
	github.com/spf13/pflag v1.0.5
	golang.org/x/lint v0.0.0-20210508222113-6edffad5e616
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	gopkg.in/fsnotify.v1 v1.4.7
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.22.2
	k8s.io/apiextensions-apiserver v0.22.2
	k8s.io/apimachinery v0.22.2
	k8s.io/client-go v0.22.2
	k8s.io/code-generator v0.22.2
	k8s.io/helm v2.16.0+incompatible
	k8s.io/klog v1.0.0
	sigs.k8s.io/controller-runtime v0.10.2
	sigs.k8s.io/controller-tools v0.7.0
)

// avoid outdated indirect dependency
replace github.com/spf13/viper => github.com/spf13/viper v1.8.1
