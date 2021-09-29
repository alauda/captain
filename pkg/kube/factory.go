package kube

import (
	"time"

	"helm.sh/helm/v3/pkg/kube"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/rest"
)

type factory struct {
	kube.Factory
}

func newFactory(f kube.Factory) kube.Factory {
	return &factory{f}
}

func adjustTimeout(req *rest.Request) {
	req.Timeout(1 * time.Minute)
}

func (f *factory) NewBuilder() *resource.Builder {
	builder := f.Factory.NewBuilder()
	return builder.TransformRequests(adjustTimeout)
}
