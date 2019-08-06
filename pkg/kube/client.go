package kube

import (
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/client-go/kubernetes/scheme"

	"helm.sh/helm/pkg/kube"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/klog"
)

func init() {
	if err := v1beta1.AddToScheme(scheme.Scheme); err != nil {
		panic(err)
	}
}

//Client is a thin wrapper around helm.kube
type Client struct {
	*kube.Client
}

// New creates a new Client.
func New(getter genericclioptions.RESTClientGetter) *Client {

	client := kube.New(getter)
	client.Factory = newFactory(client.Factory)

	return &Client{
		client,
	}
}

// Create plan to support replace.... hold on...
func (c *Client) Create(resources kube.ResourceList) (*kube.Result, error) {

	result, err := c.Client.Create(resources)
	if err != nil {
		klog.Warning("create resource error:", err)
		if errors.IsAlreadyExists(err) {
			klog.Warningf("create error due to resource exist, do a dumb update...")
			return c.Client.Update(resources, resources, true)
		}
		return result, err
	}
	return result, nil
}
