module github.com/alauda/captain

go 1.16

require (
	github.com/MakeNowJust/heredoc v1.0.0 // indirect
	github.com/Masterminds/vcs v1.13.1
	github.com/Shopify/logrus-bugsnag v0.0.0-20171204204709-577dee27f20d // indirect
	github.com/alauda/helm-crds v0.0.0-20210914035428-6e2324c2b020
	github.com/bmizerany/assert v0.0.0-20160611221934-b7ed37b82869
	github.com/bshuster-repo/logrus-logstash-hook v1.0.0 // indirect
	github.com/bugsnag/bugsnag-go v2.1.1+incompatible // indirect
	github.com/bugsnag/panicwrap v1.3.2 // indirect
	github.com/containerd/containerd v1.5.18
	github.com/containers/image/v5 v5.11.0
	github.com/davecgh/go-spew v1.1.1
	github.com/deislabs/oras v0.11.1
	github.com/docker/cli v20.10.8+incompatible // indirect
	github.com/docker/go-units v0.4.0
	github.com/garyburd/redigo v1.6.2 // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/go-logr/logr v0.4.0
	github.com/gofrs/uuid v4.0.0+incompatible // indirect
	github.com/gorilla/handlers v1.5.1 // indirect
	github.com/gosuri/uitable v0.0.4
	github.com/gsamokovarov/assert v0.0.0-20180414063448-8cd8ab63a335
	github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0 // indirect
	github.com/mattn/go-colorable v0.1.4 // indirect
	github.com/mattn/go-isatty v0.0.11 // indirect
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.13.0
	github.com/opencontainers/go-digest v1.0.0
	github.com/opencontainers/image-spec v1.0.2
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pkg/errors v0.9.1
	github.com/prometheus/common v0.26.0
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/teris-io/shortid v0.0.0-20160104014424-6c56cef5189c
	github.com/thoas/go-funk v0.5.0
	github.com/ventu-io/go-shortid v0.0.0-20170305092000-935de6796a71 // indirect
	github.com/yvasiyarov/go-metrics v0.0.0-20150112132944-c25f46c4b940 // indirect
	github.com/yvasiyarov/gorelic v0.0.7 // indirect
	github.com/yvasiyarov/newrelic_platform_go v0.0.0-20160601141957-9c099fbc30e9 // indirect
	golang.org/x/crypto v0.1.0
	gopkg.in/src-d/go-git.v4 v4.13.1
	helm.sh/helm/v3 v3.6.3
	k8s.io/api v0.21.1
	k8s.io/apiextensions-apiserver v0.21.1
	k8s.io/apimachinery v0.21.1
	k8s.io/cli-runtime v0.21.1
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/klog v1.0.0
	k8s.io/kubectl v0.21.1
	k8s.io/kubernetes v1.21.1
	sigs.k8s.io/controller-runtime v0.2.2
	sigs.k8s.io/yaml v1.2.0
)

replace (
	github.com/Masterminds/vcs => github.com/alauda/vcs v1.13.2-0.20200311111907-acd482b1ae9a
	github.com/alauda/helm-crds => github.com/alauda/helm-crds v0.0.0-20210914035428-6e2324c2b020
	github.com/deislabs/oras => github.com/deislabs/oras v0.11.0
	github.com/docker/distribution => github.com/distribution/distribution v2.7.1+incompatible
	github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309
	github.com/go-logr/logr => github.com/go-logr/logr v0.4.0
	github.com/go-logr/zapr => github.com/go-logr/zapr v0.4.0
	github.com/miekg/dns => github.com/miekg/dns v1.0.0
	github.com/xenolf/lego => github.com/go-acme/lego v0.4.0
	helm.sh/helm/v3 => github.com/alauda/helm/v3 v3.6.4-0.20210914033728-4a2cde3ea69c

	k8s.io/api => k8s.io/api v0.21.1
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.21.1
	k8s.io/apimachinery => k8s.io/apimachinery v0.21.1
	k8s.io/apiserver => k8s.io/apiserver v0.21.1
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.21.1
	k8s.io/client-go => k8s.io/client-go v0.21.1
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.21.1
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.21.1
	k8s.io/code-generator => k8s.io/code-generator v0.21.1
	k8s.io/component-base => k8s.io/component-base v0.21.1
	k8s.io/component-helpers => k8s.io/component-helpers v0.21.1
	k8s.io/controller-manager => k8s.io/controller-manager v0.21.1
	k8s.io/cri-api => k8s.io/cri-api v0.21.1
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.21.1
	k8s.io/klog/v2 => k8s.io/klog/v2 v2.0.0
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.21.1
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.21.1
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.21.1
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.21.1
	k8s.io/kubectl => k8s.io/kubectl v0.21.1
	k8s.io/kubelet => k8s.io/kubelet v0.21.1
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.21.1
	k8s.io/metrics => k8s.io/metrics v0.21.1
	k8s.io/mount-utils => k8s.io/mount-utils v0.21.1
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.21.1

	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.9.0
)
