package helm

import (
	"os"

	"github.com/alauda/captain/pkg/release/storagedriver"
	releaseclient "github.com/alauda/helm-crds/pkg/client/clientset/versioned"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/kube"
	"helm.sh/helm/v3/pkg/storage"
	"helm.sh/helm/v3/pkg/storage/driver"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
	"k8s.io/kubectl/pkg/cmd/util"
)

// getNamespace get the namespaces from ..... This may be a little unnecessary, may be we can just
// use the one we know.
func getNamespace(flags *genericclioptions.ConfigFlags) string {
	if ns, _, err := flags.ToRawKubeConfigLoader().Namespace(); err == nil {
		return ns
	}
	return "alauda-system"
}

// newActionConfig create a config for all the actions(install,delete,update...)
// allNamespaces is always set to false for now,
// default storage driver is Release now
func (d *Deploy) newActionConfig() (*action.Configuration, error) {
	restClientGetter := newConfigFlags(d.Cluster.ToRestConfig(), d.Cluster.Namespace, true)
	kubeClient := &kube.Client{
		Factory: util.NewFactory(restClientGetter),
		Log:     klog.Infof,
	}

	relClientSet, err := releaseclient.NewForConfig(d.Cluster.ToRestConfig())
	if err != nil {
		return nil, err
	}

	namespace := getNamespace(restClientGetter)
	var store *storage.Storage
	switch os.Getenv("HELM_DRIVER") {
	case "release", "releases", "":
		d := storagedriver.NewReleases(relClientSet.AppV1alpha1().Releases(namespace))
		d.Log = klog.Infof
		store = storage.Init(d)
	case "memory":
		d := driver.NewMemory()
		store = storage.Init(d)
	default:
		// Not sure what to do here.
		panic("Unknown driver in HELM_DRIVER: " + os.Getenv("HELM_DRIVER"))
	}

	d.rbacClient = &RbacClient{
		config:       d.Cluster.ToRestConfig(),
		clientGetter: restClientGetter,
	}
	d.Releases = store

	return &action.Configuration{
		RESTClientGetter: restClientGetter,
		KubeClient:       kubeClient,
		Releases:         store,
		Log:              klog.Infof,
	}, nil
}

func newConfigFlags(config *rest.Config, namespace string, insecure bool) *genericclioptions.ConfigFlags {
	return &genericclioptions.ConfigFlags{
		Namespace:   &namespace,
		APIServer:   &config.Host,
		CAFile:      &config.CAFile,
		BearerToken: &config.BearerToken,
		Insecure:    &insecure,
	}
}
