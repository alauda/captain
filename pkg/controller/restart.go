package controller

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
)

type ClusterWatchRestarter struct {
	controller *Controller
}

// NewClusterWatchRestarter ...
// This runnable means to restart invalid remote cluster connection, and make the watch working again.
func NewClusterWatchRestarter(controller *Controller) *ClusterWatchRestarter {
	return &ClusterWatchRestarter{
		controller: controller,
	}
}

// Start ...
// 1. check all the clusters to see if it's running
// 2. start watch if not for every cluster
// Limitations: If a cluster is watched already, but offline for sometime, and then back online. This runnable cannot
// handle this situation.
func (c *ClusterWatchRestarter) Start(ctx context.Context) error {
	klog.Info("start cluster restart runner...")

	// wait for the main controller
	time.Sleep(2 * time.Minute)

	return wait.PollImmediateUntil(2*time.Minute, func() (done bool, err error) {
		latest, err := c.controller.getAllClusters()
		if err != nil {
			klog.Error("get all cluster info for cluster restarter error,", err.Error())
			return false, nil
		}

		for _, item := range latest {
			cfg := item.ToRestConfig()
			coreClient, err := kubernetes.NewForConfig(cfg)
			if err != nil {
				klog.Warningf("[cluster-restarter] init core client for cluster %s error: %s", item.Name, err.Error())
				continue
			}

			version, err := coreClient.ServerVersion()
			if err != nil {
				klog.Warningf("[cluster-restarter] check cluster version for cluster %s error: %s", item.Name, err.Error())
				continue
			}
			klog.Infof("[cluster-restarter] check cluster %s version is: %s", item.Name, version.GitVersion)

			if !c.controller.IsClusterWatchStarted(item.Name) {
				if err := c.controller.restartClusterWatch(item); err != nil {
					klog.Errorf("[cluster-restarter] restart cluster watch for %s error, %s", item.Name, err.Error())
				}
				klog.Info("[cluster-restarter] restart cluster watch, ", item.Name)
			}
		}
		return false, nil
	}, ctx.Done())

}
