package controller

import (
	"context"
	"fmt"
	"strings"

	"github.com/alauda/captain/pkg/cluster"
	"github.com/alauda/captain/pkg/clusterregistry/apis/clusterregistry/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

var (
	// allClustersCacheKey is the key for ClusterList data
	allClustersCacheKey = "_all"
)

// getAllClusters list all the Clusters in global and cache it
func (c *Controller) getAllClusters() ([]*cluster.Info, error) {
	result, ok := c.ClusterCache.Get(allClustersCacheKey)
	if ok {
		return result.([]*cluster.Info), nil
	}
	klog.Infof("refresh cluster list from namespace: %s", c.clusterConfig.clusterNamespace)

	opts := v1.ListOptions{}
	var info []*cluster.Info

	list, err := c.GetClusterClient().ClusterregistryV1alpha1().Clusters(c.clusterConfig.clusterNamespace).List(opts)
	if err != nil {
		if apierrors.IsNotFound(err) {
			klog.Warning("No cluster crd found, disable multi cluster support", err)
			return info, nil

		} else {
			return nil, err
		}

	}

	for _, item := range list.Items {
		i, err := c.parseClusterInfo(&item)
		if err != nil {
			klog.Error("parse cluster info error: ", item.GetName())
		} else {
			info = append(info, i)
		}
	}
	klog.Infof("fetch %d clusters", len(info))

	c.ClusterCache.SetDefault(allClustersCacheKey, info)
	return info, nil
}

func (c *Controller) parseClusterInfo(cr *v1alpha1.Cluster) (*cluster.Info, error) {
	var info cluster.Info
	info.Name = cr.GetName()
	eps := cr.Spec.KubernetesAPIEndpoints.ServerEndpoints
	if len(eps) > 0 {
		info.Endpoint = eps[0].ServerAddress
	}

	ns := cr.Spec.AuthInfo.Controller.Namespace
	secretName := cr.Spec.AuthInfo.Controller.Name
	// get token
	sec, err := c.kubeClient.CoreV1().Secrets(ns).Get(context.Background(), secretName, v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	data, ok := sec.Data["token"]
	if ok {
		// why is there a new line.
		info.Token = strings.TrimSuffix(string(data), "\n")
		return &info, nil
	}
	return nil, fmt.Errorf(" get token error for cluster: %s", cr.Name)
}

// getClusterInfo get info about one single cluster.
// if name is "", return current cluster info
func (c *Controller) getClusterInfo(name string) (*cluster.Info, error) {
	if name == "" {
		klog.V(2).Info("find empty cluster name, use current.")
		return cluster.RestConfigToCluster(c.restConfig, cluster.DefaultClusterName), nil
	}

	data, ok := c.ClusterCache.Get(name)
	if ok {
		return data.(*cluster.Info), nil
	}

	klog.Infof("refresh cluster data: %s", name)

	options := v1.GetOptions{}
	cr, err := c.GetClusterClient().ClusterregistryV1alpha1().Clusters(c.clusterConfig.clusterNamespace).Get(name, options)
	if err != nil {
		return nil, err
	}
	info, err := c.parseClusterInfo(cr)
	if err == nil {
		c.ClusterCache.SetDefault(name, info)
	}
	return info, err
}
