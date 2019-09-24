package helm

import (
	"os"

	"github.com/alauda/captain/pkg/cluster"
	newkube "github.com/alauda/captain/pkg/kube"
	"github.com/alauda/captain/pkg/kubeconfig"
	releaseclient "github.com/alauda/helm-crds/pkg/client/clientset/versioned"
	"helm.sh/helm/pkg/kube"

	"github.com/alauda/captain/pkg/release/storagedriver"
	"helm.sh/helm/pkg/action"
	"helm.sh/helm/pkg/helmpath"

	"github.com/alauda/component-base/system"
	"helm.sh/helm/pkg/repo"
	"helm.sh/helm/pkg/storage"
	"helm.sh/helm/pkg/storage/driver"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/klog"
)

// Init do a lot of dirty stuff
// 1. init repo dir
// 2. update repo index
// we will mount the repo file from a ConfigMap by default, so this function will do the index update
func Init() {
	path := helmpath.ConfigPath("repositories.yaml")
	err := system.CreatePathIfNotExist(path)
	if err != nil {
		panic(err)
	}

	path = helmpath.CachePath("repository") + "/"
	if err := system.CreatePathIfNotExist(path); err != nil {
		panic(err)
	}

	fi, err := os.Stat(path)
	if err != nil {
		panic(err)
	} else {
		if fi.Size() > 0 {
			klog.Infof("repo file has content, update index....")
			// TODO: why this will panic
			// if err := initReposIndex(); err != nil {
			// 	panic(err)
			// }
		} else if err = repo.NewFile().WriteFile(path, 0644); err != nil {
			panic(err)
		}
	}
}

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
func newActionConfig(info *cluster.Info) (*action.Configuration, error) {
	cfg, err := kubeconfig.UpdateKubeConfig(info)
	if err != nil {
		return nil, err
	}
	cfg.Namespace = info.Namespace

	cfgFlags := kube.GetConfig(cfg.Path, cfg.Context, cfg.Namespace)
	kc := newkube.New(cfgFlags)
	// hope it works
	kc.Log = klog.Infof

	namespace := getNamespace(cfgFlags)

	relClientSet, err := releaseclient.NewForConfig(info.ToRestConfig())
	if err != nil {
		return nil, err
	}

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

	return &action.Configuration{
		RESTClientGetter: cfgFlags,
		KubeClient:       kc,
		Releases:         store,
		Log:              klog.Infof,
	}, nil
}


