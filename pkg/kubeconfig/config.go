package kubeconfig

import (
	"github.com/alauda/captain/pkg/cluster"
	"github.com/alauda/component-base/system"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/klog"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/kubeconfig"
)

var (
	// defaultPath is the kuebconfig path
	defaultPath = ".kube/config"
)

//Config ...
type Config struct {
	Context   string
	Namespace string
	Path      string
}

func createKubeConfig(info *cluster.Info) *clientcmdapi.Config {
	cfg := kubeconfig.CreateWithToken(info.Endpoint, info.Name, info.Name, nil, info.Token)
	cfg.Clusters[info.Name].InsecureSkipTLSVerify = true
	return cfg
}

// isContextChanged check if a context for cluster has changed, like endpoint, token....
func isContextChanged(new *clientcmdapi.Config, old *clientcmdapi.Config) bool {
	for k, v := range new.Clusters {
		if old.Clusters[k].Server != v.Server {
			return true
		}
	}

	for k, v := range new.AuthInfos {
		if old.AuthInfos[k].Token != v.Token {
			return true
		}
	}

	return false

}

func mergeKubeConfig(new *clientcmdapi.Config, old *clientcmdapi.Config) {
	for k, v := range new.Clusters {
		old.Clusters[k] = v
	}
	for k, v := range new.Contexts {
		old.Contexts[k] = v
	}
	for k, v := range new.AuthInfos {
		old.AuthInfos[k] = v
	}
}

// UpdateKubeConfig add a context to kubeconfig file for one cluster,
// we don't set namespace so we can reuse it
func UpdateKubeConfig(info *cluster.Info) (*Config, error) {
	cfg := Config{
		Path:      defaultPath,
		Context:   info.GetContext(),
		Namespace: "alauda-system",
	}

	if err := system.CreatePathIfNotExist(cfg.Path); err != nil {
		return nil, err
	}

	kubeConfig, err := clientcmd.LoadFromFile(cfg.Path)
	if err != nil {
		return nil, err
	}

	newKubeConfig := createKubeConfig(info)

	_, ok := kubeConfig.Contexts[info.GetContext()]
	if ok && !isContextChanged(newKubeConfig, kubeConfig) {
		klog.Infof("context %s already exist", info.GetContext())
		return &cfg, nil
	}

	klog.Infof("create or update context for: %s", info.GetContext())

	mergeKubeConfig(newKubeConfig, kubeConfig)
	if err := kubeconfig.WriteToDisk(cfg.Path, kubeConfig); err != nil {
		return nil, err
	}
	return &cfg, nil
}
