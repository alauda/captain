package helm

import (
	"strings"
	"time"

	"github.com/alauda/captain/pkg/util"
	clientset "github.com/alauda/helm-crds/pkg/client/clientset/versioned"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/storage/driver"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kblabels "k8s.io/apimachinery/pkg/labels"
)

// Delete delete a Release from a cluster
func (d *Deploy) Delete() error {
	hr := d.HelmRequest

	name := GetReleaseName(hr)

	cfg, err := d.newActionConfig()
	if err != nil {
		return err
	}

	client := action.NewUninstall(cfg)
	client.Timeout = 60 * time.Second

	client.KeepResources = isSwitchEnabled(hr, util.KeepResourcesAnnotation)
	if client.KeepResources {
		d.Log.Info("found keep resources in helmrequest annotation, will keep k8s resources when uninstall current release ", "name", name)
	}

	res, err := client.Run(name)
	if err != nil {
		d.Log.Error(err, "uninstall error", "name", name)
		if errors.Cause(err) == driver.ErrReleaseNotFound {
			d.Log.Info("release not exist when delete, ignore it", "name", name)
			return nil
		}

		// if we cannot access the target cluster, the helmrequest should be able to be deleted
		if strings.HasSuffix(err.Error(), "EOF") || strings.Contains(err.Error(), "connect: connection refused") {
			d.Log.Info("target cluster cannot access when delete helmrequest", "cluster", hr.Spec.ClusterName, "name", hr.GetName())
			return nil
		}

		// if we have object not found error, it may be stuck at `uninstalling`, we can handle this
		if strings.Contains(err.Error(), "object not found, skipping delete") {
			if err := d.forceDeleteRelease(); err != nil {
				d.Log.Error(err, "force delete stuck release error")
				return err
			}
			return nil
		}

		return err
	}
	if res != nil && res.Info != "" {
		d.Log.Info("release uninstalled", "name", name, "info", res.Info)
	} else {
		d.Log.Info("release uninstalled", "name", name)
	}
	return nil
}

// if something block the deletion, and we think it can be ignored, we can do a force delete,
// remove the `uninstalling` release. This can be caused by
// 1. unable to build resources(TODO: move from main)
// 2. object not found
func (d *Deploy) forceDeleteRelease() error {
	name := d.HelmRequest.Name
	client, err := clientset.NewForConfig(d.Cluster.ToRestConfig())
	if err != nil {
		d.Log.Error(err, "create client error when handle error on delete", "name", name)
		return err
	}
	options := metav1.ListOptions{
		LabelSelector: kblabels.Set{
			"name":   d.HelmRequest.GetName(),
			"status": "uninstalling",
		}.AsSelector().String(),
	}
	hist, err := client.AppV1alpha1().Releases(d.HelmRequest.Namespace).List(options)
	if err != nil {
		d.Log.Error(err, "unable to list uninstalling releases", "name", name)
	} else {
		return err
	}

	if len(hist.Items) > 0 {
		for _, item := range hist.Items {
			d.Log.Info("deleting stuck uninstalling release", "name", item.Name)
			if err := client.AppV1alpha1().Releases(d.HelmRequest.Namespace).Delete(item.Name, &metav1.DeleteOptions{}); err != nil {
				d.Log.Error(err, "delete uninstalling release error", "name", item.Name)
			} else {
				return err
			}
		}
	}
	return nil

}
