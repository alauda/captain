package controller

import (
	"fmt"
	"os"

	"github.com/alauda/captain/pkg/apis/app/v1alpha1"
	"github.com/alauda/captain/pkg/cluster"
	"github.com/alauda/captain/pkg/helm"
	"github.com/alauda/captain/pkg/release"
	funk "github.com/thoas/go-funk"
	"helm.sh/helm/pkg/action"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/klog"
)

// syncToAllClusters install/upgrade release in all the clusters
func (c *Controller) syncToAllClusters(key string, helmRequest *v1alpha1.HelmRequest) error {
	clusters, err := c.getAllClusters()
	if err != nil {
		return err
	}

	var synced []string
	var errs []error
	equal := helm.IsHelmRequestSynced(helmRequest)

	// if not equal, we need to update helm status first
	if !equal {
		helmRequest.Status.SyncedClusters = make([]string, 0)
		if err := c.updateHelmRequestPhase(helmRequest, v1alpha1.HelmRequestPending); err != nil {
			return err
		}
	} else if helmRequest.Status.SyncedClusters != nil {
		// if hash equal, record synced clusters
		synced = helmRequest.Status.SyncedClusters
	}
	klog.Infof("origin synced clusters: %+v", synced)

	for _, cr := range clusters {
		if equal && funk.Contains(synced, cr.Name) {
			continue
		}
		klog.Infof("sync %s to cluster %s ....", key, cr.Name)
		if err = c.sync(cr, helmRequest); err != nil {
			errs = append(errs, err)
			klog.Infof("skip sync %s to %s, err is : %s, continue...", key, cr.Name, err.Error())
			continue
		}
		// avoid duplicates...
		if !funk.Contains(synced, cr.Name) {
			synced = append(synced, cr.Name)
		}
	}

	helmRequest.Status.SyncedClusters = synced
	klog.Infof("synced %s to clusters: %+v", key, synced)

	err = errors.NewAggregate(errs)

	if len(synced) >= len(clusters) {
		// all synced
		return c.updateHelmRequestStatus(helmRequest)
	} else if len(synced) > 0 {
		// partial synced
		c.sendFailedSyncEvent(helmRequest, err)
		return c.setPartialSyncedStatus(helmRequest)
	}
	return err
}

// sync install/update chart to one cluster
func (c *Controller) sync(info *cluster.Info, helmRequest *v1alpha1.HelmRequest) error {
	info.Namespace = helmRequest.Spec.Namespace
	if err := release.EnsureCRDCreated(info.ToRestConfig()); err != nil {
		klog.Errorf("sync release crd error: %s", err.Error())
		return err
	}

	rel, err := helm.Sync(helmRequest, info)
	if err != nil {
		return err
	}

	// record chart version for un-specified ones
	msg := fmt.Sprintf("Choose chart version: %s %s", rel.Chart.Metadata.Name, rel.Chart.Metadata.Version)
	c.recorder.Event(helmRequest, corev1.EventTypeNormal, SuccessSynced, msg)

	action.PrintRelease(os.Stdout, rel)
	return nil
}

// syncToCluster install/update HelmRequest to one cluster
func (c *Controller) syncToCluster(helmRequest *v1alpha1.HelmRequest) error {
	clusterName := helmRequest.Spec.ClusterName
	info, err := c.getClusterInfo(clusterName)
	if err != nil {
		klog.Errorf("get cluster info error: %s", err.Error())
		return err
	}

	klog.Infof("get cluster %s  endpoint: %s", info.Name, info.Endpoint)

	if err := c.sync(info, helmRequest); err != nil {
		return err
	}

	// Finally, we update the status block of the HelmRequest resource to reflect the
	// current state of the world
	err = c.updateHelmRequestStatus(helmRequest)
	return err
}
