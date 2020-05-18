package controller

import (
	"fmt"
	"github.com/pkg/errors"
	"os"

	"github.com/alauda/captain/pkg/cluster"
	"github.com/alauda/captain/pkg/helm"
	"github.com/alauda/captain/pkg/release"
	"github.com/alauda/helm-crds/pkg/apis/app/v1beta1"
	"github.com/thoas/go-funk"
	"helm.sh/helm/pkg/action"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kblabels "k8s.io/apimachinery/pkg/labels"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/klog"
)

// syncToAllClusters install/upgrade release in all the clusters
func (c *Controller) syncToAllClusters(key string, helmRequest *v1beta1.HelmRequest) error {
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
		if err := c.updateHelmRequestPhase(helmRequest, v1beta1.HelmRequestPending); err != nil {
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

	err = utilerrors.NewAggregate(errs)

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
func (c *Controller) sync(info *cluster.Info, helmRequest *v1beta1.HelmRequest) error {
	ci := *info
	ci.Namespace = helmRequest.Spec.Namespace
	if err := release.EnsureCRDCreated(info.ToRestConfig()); err != nil {
		klog.Errorf("sync release crd error: %s", err.Error())
		return err
	}

	deploy := helm.NewDeploy()

	// found exist release here, this is logic from helm, and we skip the decode part to
	// avoid OOM. This may be removed in the feature
	// TODO: may be a bug ,if installToAllCluster, may be get the wrong release.
	client := c.getAppClientForRelease(helmRequest)
	if client == nil {
		// may be not inited yet
		err := errors.New(fmt.Sprintf("get client for release error, retry later. cluster is: %s", helmRequest.Spec.ClusterName))
		return err
	}
	options := metav1.ListOptions{
		LabelSelector: kblabels.Set{"name": helmRequest.GetName()}.AsSelector().String(),
	}
	hist, err := client.AppV1alpha1().Releases(helmRequest.Spec.Namespace).List(options)
	deployed := false
	if err == nil && len(hist.Items) > 0 {
		// deployed = true
		for _, item := range hist.Items {
			if item.Status.Status == "deployed" {
				deployed = true
			}

			// delete pending-install releases, may be caused by OOM
			if item.Status.Status == "pending-install" || item.Status.Status == "uninstalling" || item.Status.Status == "pending-upgrade" || item.Status.Status == "failed" {
				deploy.Log.Info("found pending release, planning to delete it", "name", item.Name, "status", item.Status.Status)
				if err := client.AppV1alpha1().Releases(helmRequest.Spec.Namespace).Delete(item.Name, &metav1.DeleteOptions{}); err != nil {
					deploy.Log.Error(err, "delete pending release error", "name", item.Name)
				}
			}
		}
	}

	inCluster, _ := c.getClusterInfo("")

	deploy.Deployed = deployed
	deploy.Cluster = &ci
	deploy.InCluster = inCluster
	deploy.SystemNamespace = c.systemNamespace
	deploy.HelmRequest = helmRequest

	rel, err := deploy.Sync()
	if err != nil {
		return err
	}

	// record chart version for un-specified ones
	msg := fmt.Sprintf("Choose chart version: %s %s", rel.Chart.Metadata.Name, rel.Chart.Metadata.Version)
	c.getEventRecorder(helmRequest).Event(helmRequest, corev1.EventTypeNormal, SuccessSynced, msg)

	action.PrintRelease(os.Stdout, rel)
	return nil
}
