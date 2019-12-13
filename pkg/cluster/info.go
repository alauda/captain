package cluster

import (
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/cluster-registry/pkg/apis/clusterregistry/v1alpha1"
	"k8s.io/cluster-registry/pkg/client/clientset/versioned"
	"k8s.io/klog"
)

const (
	//DefaultClusterName is the cluster for the unspecified cluster
	DefaultClusterName = "_default"
)

//Info represents a Cluster,
type Info struct {
	// Name is the cluster name, usually the Cluster resource's name
	Name string
	// Endpoint the apiserver's endpoint of the cluster
	Endpoint string
	// Token is a admin token , it should have all the access to the cluster
	Token string

	// Namespace the namespace which the chart will be installed to
	Namespace string
}

//GetContext is the context name for this cluster, this name format is generated from k8s code
func (i *Info) GetContext() string {
	return i.Name + "@" + i.Name
}

//ToRestConfig generate rest.Config from cluster info.
func (i *Info) ToRestConfig() *rest.Config {
	return &rest.Config{
		Host:        i.Endpoint,
		BearerToken: i.Token,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: true,
		},
	}
}

//RestConfigToCluster generate a cluster Info from a rest config
// This method and the Info.ToRestConfig both only support bearer token for now
// luckily, the in-cluster rest config also use bearer token
func RestConfigToCluster(config *rest.Config, generatedName string) *Info {
	var i Info
	i.Token = config.BearerToken
	i.Endpoint = config.Host
	i.Name = generatedName
	return &i
}

// GetClusters get clusters resourcese. If this resource not exist on cluster, just ignore it.
func GetClusters(client versioned.Interface, ns string, opts metav1.ListOptions) (*v1alpha1.ClusterList, error) {
	origin, err := client.ClusterregistryV1alpha1().Clusters(ns).List(opts)
	if err != nil {
		if apierrors.IsNotFound(err) {
			klog.Warning("no cluster found:", err)
			return origin, nil
		}
	}
	return origin, err
}
