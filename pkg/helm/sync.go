package helm

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/alauda/captain/pkg/cluster"
	"github.com/alauda/captain/pkg/helmrequest"
	"github.com/alauda/captain/pkg/util"
	"github.com/alauda/helm-crds/pkg/apis/app/v1alpha1"
	clientset "github.com/alauda/helm-crds/pkg/client/clientset/versioned"
	"github.com/go-logr/logr"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"github.com/teris-io/shortid"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
)

// Deploy contains info about one chart deploy
type Deploy struct {
	// logger
	Log logr.Logger

	// is this chart has deployed release
	Deployed bool

	// the global cluster info
	InCluster *cluster.Info

	// target cluster info
	Cluster *cluster.Info

	// system namespace for chartrepo
	SystemNamespace string

	// all the charts info
	HelmRequest *v1alpha1.HelmRequest

	// Client is the crd client for hr
	Client clientset.Interface

	rbacClient *RbacClient

	// Releases stores records of releases.
	Releases *storage.Storage
}

type RbacClient struct {
	// config is the rest config for target cluster, use it as base and add Impersonate info for rbac check
	config *rest.Config

	// verb is the rbac verb we want to check. For chart install/update/delete, we want to check create/update/delete verb.
	// this is a simple assumption, since create may contains craete/update/delete action...
	verb string

	// clientGetter is for helm kube client
	clientGetter genericclioptions.RESTClientGetter
}

// NewDeploy create a new deploy struct with crd client set
func NewDeploy(client clientset.Interface) *Deploy {
	uid, _ := shortid.Generate()
	log := ctrl.Log.WithName("helm-" + strings.ToLower(uid))

	var d Deploy
	d.Log = log
	d.Client = client
	return &d
}

// we need to save the charts to cache to avoid repeat download
var chartCache = cache.New(30*time.Minute, 60*time.Minute)

func (d *Deploy) GetCurrentReleases() ([]*release.Release, error) {
	name := GetReleaseName(d.HelmRequest)
	return d.Releases.History(name)
}

// Sync = install + upgrade
// When sync done, add the release note to HelmRequest status
// inCluster info is used to retrieve config info for valuesFrom
func (d *Deploy) Sync() (*release.Release, error) {
	log := d.Log
	hr := d.HelmRequest

	name := GetReleaseName(hr)
	out := os.Stdout

	// helm settings
	settings := cli.New()
	settings.Debug = true

	// init upgrade client
	cfg, err := d.newActionConfig()
	if err != nil {
		return nil, err
	}
	client := action.NewUpgrade(cfg)
	// client.Force = true
	client.Namespace = hr.GetReleaseNamespace()
	client.Install = true
	// This should be a reasonable value
	client.MaxHistory = 10
	// Do not validate openapi schema
	client.DisableOpenAPIValidation = true
	// Timeout as the same as install
	client.Timeout = 180 * time.Second
	// client.ForceAdopt = true
	client.ForceAdopt = isSwitchEnabled(hr, util.ForceAdoptResourcesAnnotation)

	if hr.Spec.Version != "" {
		client.Version = hr.Spec.Version
	}

	// merge values
	values, err := getValues(hr, d.InCluster.ToRestConfig())
	if err != nil {
		return nil, err
	}
	client.ResetValues = true

	// load from cache first, then from disk
	var ch *chart.Chart
	downloader := NewDownloader(d.SystemNamespace, d.InCluster.ToRestConfig(), d.Cluster.ToRestConfig(), d.Log)
	chartPath, err := downloader.downloadChart(hr.Spec.Chart, helmrequest.ResolveVersion(hr))
	if err != nil {
		return nil, err
	}
	log.Info("load charts from disk", "path", chartPath)
	ch, err = loader.Load(chartPath)
	if err != nil {
		log.Error(err, "failed to load chart from disk ", "path", chartPath)
		return nil, err
	}
	d.HelmRequest.Status.Version = ch.Metadata.Version

	chname := fmt.Sprintf("%s:%s", ch.Metadata.Name, ch.Metadata.Version)
	if err := d.setChartLoadedCondition(chname); err != nil {
		log.Error(err, "set chart downloaded condition error")
	}

	if req := ch.Metadata.Dependencies; req != nil {
		if err := action.CheckDependencies(ch, req); err != nil {
			return nil, err
		}
	}

	if !d.Deployed {
		log.Info("Release does not exist. Installing it now", "name", name)
		resp, err := d.install(ch)
		if err != nil {
			if !strings.Contains(err.Error(), "cannot re-use a name that is still in use") {
				// if error occurred, just return. Otherwise the upgrade will stuck at no deploy found
				log.Error(err, "install before upgrade failed", "name", hr.Name)
				return resp, err
			}
			log.Info("Release maybe already exists when install it. Will upgrade it.", "name", name)
		} else {
			hr.Status.Notes = resp.Info.Notes
			return resp, nil
		}
	}

	// run upgrade/install
	resp, err := client.Run(name, ch, values)
	if err != nil {
		return nil, errors.Wrap(err, "UPGRADE FAILED")
	}
	PrintRelease(out, resp)
	log.Info("Release has been upgraded. Happy Helming!\n", "name", name)

	// Print the status like status command does
	statusClient := action.NewStatus(cfg)
	rel, err := statusClient.Run(name)
	if err != nil {
		log.Error(err, "print status error")

	}
	PrintRelease(out, rel)
	if rel != nil {
		hr.Status.Notes = rel.Info.Notes
	}
	return resp, nil

}
