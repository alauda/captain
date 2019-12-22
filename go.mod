module github.com/alauda/captain

go 1.13

replace helm.sh/helm => github.com/alauda/helm v3.0.0-beta.3.0.20191221113435-3710e11fac68+incompatible

replace github.com/deislabs/oras => github.com/deislabs/oras v0.7.0

replace github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309

require k8s.io/kubernetes v1.15.0

replace (
	k8s.io/api => k8s.io/api v0.0.0-20190620084959-7cf5895f2711
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190620085554-14e95df34f1f
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190612205821-1799e75a0719
	k8s.io/apiserver => k8s.io/apiserver v0.0.0-20190620085212-47dc9a115b18
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20190620085706-2090e6d8f84c
	k8s.io/client-go => k8s.io/client-go v0.0.0-20190620085101-78d2af792bab
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.0.0-20190620090043-8301c0bda1f0
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.0.0-20190620090013-c9a0fc045dc1
	k8s.io/code-generator => k8s.io/code-generator v0.0.0-20190612205613-18da4a14b22b
	k8s.io/component-base => k8s.io/component-base v0.0.0-20190620085130-185d68e6e6ea
	k8s.io/cri-api => k8s.io/cri-api v0.0.0-20190531030430-6117653b35f1
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.0.0-20190620090116-299a7b270edc
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.0.0-20190620085325-f29e2b4a4f84
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.0.0-20190620085942-b7f18460b210
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.0.0-20190620085809-589f994ddf7f
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.0.0-20190620085912-4acac5405ec6
	k8s.io/kubelet => k8s.io/kubelet v0.0.0-20190620085838-f1cb295a73c9
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.0.0-20190620090156-2138f2c9de18
	k8s.io/metrics => k8s.io/metrics v0.0.0-20190620085625-3b22d835f165
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.0.0-20190620085408-1aef9010884e
)

require (
	github.com/Jeffail/gabs/v2 v2.3.0
	github.com/MakeNowJust/heredoc v1.0.0 // indirect
	github.com/Masterminds/goutils v1.1.0 // indirect
	github.com/Masterminds/sprig v2.22.0+incompatible // indirect
	github.com/alauda/component-base v0.0.0-20190628064654-a4dafcfd3446
	github.com/alauda/helm-crds v0.0.0-20191219064609-456571553411
	github.com/containerd/containerd v1.3.2 // indirect
	github.com/coreos/bbolt v1.3.3 // indirect
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f // indirect
	github.com/deislabs/oras v0.8.0 // indirect
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/fatih/camelcase v1.0.0 // indirect
	github.com/fatih/color v1.7.0 // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/go-logr/logr v0.1.0
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/gorilla/websocket v1.4.1 // indirect
	github.com/gosuri/uitable v0.0.4 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.1.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.12.1 // indirect
	github.com/gsamokovarov/assert v0.0.0-20180414063448-8cd8ab63a335
	github.com/huandu/xstrings v1.2.1 // indirect
	github.com/mattn/go-colorable v0.1.4 // indirect
	github.com/mattn/go-isatty v0.0.11 // indirect
	github.com/mattn/go-runewidth v0.0.7 // indirect
	github.com/mitchellh/copystructure v1.0.0 // indirect
	github.com/mitchellh/go-wordwrap v1.0.0 // indirect
	github.com/onsi/ginkgo v1.8.0
	github.com/onsi/gomega v1.5.0
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pkg/errors v0.8.1
	github.com/soheilhy/cmux v0.1.4 // indirect
	github.com/thoas/go-funk v0.5.0
	github.com/tmc/grpc-websocket-proxy v0.0.0-20190109142713-0ad062ec5ee5 // indirect
	github.com/xiang90/probing v0.0.0-20190116061207-43a291ad63a2 // indirect
	go.etcd.io/bbolt v1.3.3 // indirect
	gopkg.in/square/go-jose.v2 v2.4.1 // indirect
	helm.sh/helm v3.0.0-alpha.1.0.20190613170622-c35dbb7aabf8+incompatible
	k8s.io/api v0.16.4
	k8s.io/apiextensions-apiserver v0.0.0
	k8s.io/apimachinery v0.17.0
	k8s.io/cli-runtime v0.16.4
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/cluster-registry v0.0.6
	k8s.io/klog v1.0.0
	rsc.io/letsencrypt v0.0.3 // indirect
	sigs.k8s.io/controller-runtime v0.2.2
)
