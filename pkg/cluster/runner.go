package cluster

import (
	"os"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	clusterclientset "k8s.io/cluster-registry/pkg/client/clientset/versioned"
	"k8s.io/klog"
)

type ClusterRefresher struct {
	cfg *rest.Config
	ns  string
}

// NewClusterRefresher ...
// This runnable intent to inform captain to restart when it found a new cluster. This is not a ideal solution, but it
// works on most occasions. Add/Remove cluster should be a rear operation in prod environment.
func NewClusterRefresher(ns string, cfg *rest.Config) *ClusterRefresher {
	return &ClusterRefresher{
		cfg: cfg,
		ns:  ns,
	}
}

func (c *ClusterRefresher) Start(stopCh <-chan struct{}) error {
	klog.Info("start cluster refresher runner...")

	clusterClient, err := clusterclientset.NewForConfig(c.cfg)
	if err != nil {
		klog.Info("init cluster client error")
		return err
	}

	opts := metav1.ListOptions{}

	origin, err := GetClusters(clusterClient, c.ns, opts)
	if err != nil {
		return err
	}

	for {
		// short enough to avoid re-install kubernetes cluster on the same cluster.
		time.Sleep(15 * time.Second)

		latest, err := GetClusters(clusterClient, c.ns, opts)
		if err != nil {
			return err
		}

		// avoid re-install kubernetes cluster on the same nodes. Restart twice should work
		if len(latest.Items) != len(origin.Items) {
			klog.Info("possible new cluster added, restart captain")
			time.Sleep(60 * time.Second)
			os.Exit(0)
		}

	}

}
