package kube

import (
	"bytes"
	"io"
	"io/ioutil"

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

// BuildUnstructured  validates for Kubernetes objects and returns unstructured infos.
// Maybe this override is not needed anymore
func (c *Client) BuildUnstructured(reader io.Reader) (kube.Result, error) {
	result, err := c.Client.BuildUnstructured(reader)
	if err != nil {
		klog.Warning("build unstructured error: ", err)
		return result, err

		// if strings.Contains(err.Error(), "apiVersion") && strings.Contains(err.Error(), "is not available") {
		//	klog.Warning("encountered apiVersion not found it, ignore it: ", err)
		//	return result, nil
		// }
	}
	return result, err
}

// Build validates for Kubernetes objects and returns resource Infos from a io.Reader.
func (c *Client) Build(reader io.Reader) (kube.Result, error) {
	return c.Client.BuildUnstructured(reader)
}

// Create plan to support replace.... hold on...
func (c *Client) Create(reader io.Reader) error {
	buf, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}

	create := bytes.NewBuffer(buf)
	origin := bytes.NewBuffer(buf)
	target := bytes.NewBuffer(buf)

	if err := c.Client.Create(create); err != nil {
		klog.Warning("create resource error:", err)
		if errors.IsAlreadyExists(err) {
			klog.Warningf("create error due to resource exist, do a dumb update...")
			return c.Client.Update(origin, target, true, false)
		}
		return err
	}
	return nil
}
