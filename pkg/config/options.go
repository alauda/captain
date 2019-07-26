package config

import (
	"flag"

	"alauda.io/captain/pkg/util"

	"sigs.k8s.io/controller-runtime/pkg/manager"
)

//Options contains all the options for captain
type Options struct {
	// Options contains some useful options
	manager.Options

	// MasterURL is the url of kubernetes apiserver
	MasterURL string
	// KubeConfig is the path for kubeconfig file
	KubeConfig string
	// InstallCRD determine if we should install the HelmRequest CRD in the controller
	InstallCRD bool
	// ClusterNamespace is the namespace where all the Cluster resources lives in
	ClusterNamespace string

	// EnableValidateWebhook decide if we should enable the validating webhook
	// mainly used for local test
	EnableValidateWebhook bool

	// GlobalClusterName it the cluster'name who represents the global cluster, which is also the
	// cluster captain lives in. We need this variable because we want to support dependency of a
	// HelmRequest who's clusterName="", which should means the current cluster. Currently this
	// seems the only options, we can't compare them through endpoint or token ( not equal when captain
	// run in-cluster mode). Hope there will be a better way in the feature.
	GlobalClusterName string

	// PrintVersion print the version and exist
	PrintVersion bool
}

func (opt *Options) setDefaults() {
	opt.LeaderElectionID = util.LeaderLockName
}

// BindFlags init flags and options
func (opt *Options) BindFlags() {
	opt.setDefaults()

	flag.StringVar(&opt.KubeConfig, "kubeconfig", "",
		"Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&opt.MasterURL, "master", "",
		"The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	flag.BoolVar(&opt.PrintVersion, "version", false,
		"Print version")
	flag.BoolVar(&opt.InstallCRD, "install-crd", true,
		"Install HelmRequest CRD if it does not exist")
	flag.StringVar(&opt.ClusterNamespace, "cluster-namespace", "alauda-system",
		"The namespace where all the Cluster resource lives in")
	flag.StringVar(&opt.GlobalClusterName, "global-cluster-name", "global",
		"The name of the global cluster resource")
	// EnableLeaderElection decide if we should enable leader election
	// this flag is mainly used to enable local test. If enabled, the controller will also
	// do a simple check to see if it's running in a kubernetes cluster. If passed,
	// then the leader election is finally turned on
	flag.BoolVar(&opt.LeaderElection, "enable-leader-election", true,
		"Enable leader election")
	flag.BoolVar(&opt.EnableValidateWebhook, "enable-validating-webhook", true,
		"Enable validating webhook")

	flag.StringVar(&opt.MetricsBindAddress, "metrics-bind-address", ":6060",
		"Setup bind address for metrics server, use \"\" to disable it")

}
