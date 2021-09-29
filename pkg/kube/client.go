package kube

import (
	"io"
	"sync"
	"time"

	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/kube"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
)

// use lock for protect client generation
var mu sync.Mutex

func init() {
	if err := v1beta1.AddToScheme(scheme.Scheme); err != nil {
		panic(err)
	}
}

//Client is a thin wrapper around helm.kube
type Client struct {
	*kube.Client

	core kubernetes.Interface
}

// New creates a new Client.
func New(getter genericclioptions.RESTClientGetter, config *rest.Config) *Client {
	mu.Lock()
	client := kube.New(getter)
	mu.Unlock()
	client.Factory = newFactory(client.Factory)

	core := kubernetes.NewForConfigOrDie(config)

	return &Client{
		Client: client,
		core:   core,
	}
}

// Build validates for Kubernetes objects and returns resource Infos from a io.Reader.
func (c *Client) Build(reader io.Reader, validate bool) (kube.ResourceList, error) {
	result, err := c.Client.Build(reader, validate)
	if err != nil {
		klog.Warning("build resources error: ", err)
		return result, err

		// if strings.Contains(err.Error(), "apiVersion") && strings.Contains(err.Error(), "is not available") {
		//	klog.Warning("encountered apiVersion not found it, ignore it: ", err)
		//	return result, nil
		// }
	}
	return result, nil
}

// Create plan to support replace.... hold on...
func (c *Client) Create(resources kube.ResourceList) (*kube.Result, error) {

	result, err := c.Client.Create(resources)
	if err != nil {
		klog.Warning("create resource error:", err)
		if apierrors.IsAlreadyExists(err) {
			klog.Warningf("create error due to resource exist, do a dumb update...")
			// result, err := c.Client.Update(resources, resources, true)
			return c.timeoutUpdate(resources)
		}
		return result, err
	}
	return result, nil
}

func (c *Client) timeoutUpdate(resources kube.ResourceList) (*kube.Result, error) {
	type res struct {
		result *kube.Result
		err    error
	}
	c1 := make(chan res, 1)
	go func() {
		result, err := c.Client.Update(resources, resources, true)
		c1 <- res{
			result: result,
			err:    err,
		}
	}()

	select {
	case res := <-c1:
		return res.result, res.err
	case <-time.After(10 * time.Second):
		klog.Warningf("timeout create-update resource: %s, ignore it", resources[0].Name)
		return nil, nil
	}
}

func (c *Client) IsReachable() error {
	client, err := c.Factory.KubernetesClientSet()
	if err != nil {
		klog.Error("create kubernetes client error when check cluster is reachable: ", err)
		return err
	}
	_, err = client.ServerVersion()
	if err != nil {
		return errors.New("Kubernetes cluster unreachable")
	}
	return nil
}
