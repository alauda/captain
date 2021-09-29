package controller

import (
	"fmt"

	"github.com/alauda/helm-crds/pkg/apis/app/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

// getHelmRequestDependencies get dependencies for a HelmRequest resource
// If the target HelmRequest has no dependencies, return nil. Otherwise get the dependencies and return
func (c *Controller) getHelmRequestDependencies(hr *v1alpha1.HelmRequest) ([]*v1alpha1.HelmRequest, error) {
	var data []*v1alpha1.HelmRequest
	deps := hr.Spec.Dependencies
	if len(deps) == 0 {
		klog.V(4).Infof("HelmRequest %s has no dependencies", hr.GetName())
		return nil, nil
	}

	cluster := c.getDeployCluster(hr)
	for _, name := range deps {
		d, err := c.getClusterAppClient(cluster).AppV1alpha1().HelmRequests(hr.GetNamespace()).Get(name, metav1.GetOptions{})
		if err != nil {
			klog.Errorf("Retrieve dependency %s for %s error: %s", name, hr.GetName(), err.Error())
			return nil, err
		}
		data = append(data, d)
	}

	return data, nil

}

// checkDependenciesForHelmRequest checks if the dependencies for the target HelmRequest has been
// satisfied. This is not and easy job, since we support installToAllClusters, and the clusters live and
// go all the time. For more details please check: http://confluence.alaudatech.com/pages/viewpage.action?pageId=48729300
// If the check not pass or somethings goes wrong, return an error contains the detailed reson, this is
// better than a bool var.
func (c *Controller) checkDependenciesForHelmRequest(hr *v1alpha1.HelmRequest) error {
	deps, err := c.getHelmRequestDependencies(hr)
	if err != nil {
		return err
	}

	if !hr.Spec.InstallToAllClusters {
		cluster := c.getDeployCluster(hr)
		for _, dep := range deps {
			if !dep.IsClusterSynced(cluster) && (cluster == "" && !dep.IsClusterSynced(c.clusterConfig.globalClusterName)) {
				return fmt.Errorf("dependency %s of %s is not synced to cluster %s yet", dep.Name, hr.Name, cluster)
			}
		}
		return nil
	}

	clusters, err := c.getAllClusters()
	if err != nil {
		return fmt.Errorf("get clusters info error when check dependencies for %s : %s", hr.Name, err.Error())
	}

	for _, item := range clusters {
		for _, dep := range deps {
			if !dep.IsClusterSynced(item.Name) {
				return fmt.Errorf("dependency %s of %s is not synced to cluster %s yet", dep.Name, hr.Name, item.Name)
			}
		}
	}
	return nil

}
