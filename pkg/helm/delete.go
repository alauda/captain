package helm

import (
	"github.com/alauda/captain/pkg/apis/app/v1alpha1"
	"github.com/alauda/captain/pkg/cluster"
	"github.com/pkg/errors"
	"helm.sh/helm/pkg/action"
	"helm.sh/helm/pkg/storage/driver"
	"k8s.io/klog"
)

// Delete delete a Release from a cluster
func Delete(hr *v1alpha1.HelmRequest, info *cluster.Info) error {

	name := getReleaseName(hr)

	cfg, err := newActionConfig(info)
	if err != nil {
		return err
	}

	client := action.NewUninstall(cfg)

	res, err := client.Run(name)
	if err != nil {
		if errors.Cause(err) == driver.ErrReleaseNotFound {
			klog.Warning("release not exist when delete, ignore it:", name)
			return nil
		}
		return err
	}
	if res != nil && res.Info != "" {
		klog.Infof("release %s uninstalled, info: %+v", name, res.Info)
	} else {
		klog.Infof("release \"%s\" uninstalled\n", name)
	}
	return nil
}
