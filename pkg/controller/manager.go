package controller

import (
	appscheme "github.com/alauda/helm-crds/pkg/client/clientset/versioned/scheme"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/scheme"
)

func init() {
	// Add sample-controller types to the default Kubernetes Scheme so Events can be
	// logged for sample-controller types.
	runtime.Must(appscheme.AddToScheme(scheme.Scheme))
}

// NeedLeaderElection simply check token file to determine it's this a in-cluster config and enable
// leader-election
func (c *Controller) NeedLeaderElection() bool {
	return c.restConfig.BearerTokenFile != ""
}
