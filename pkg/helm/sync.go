package helm

import (
	"github.com/alauda/captain/pkg/cluster"
	"github.com/alauda/helm-crds/pkg/apis/app/v1alpha1"
	"github.com/go-logr/logr"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"github.com/teris-io/shortid"
	"helm.sh/helm/pkg/action"
	"helm.sh/helm/pkg/chart"
	"helm.sh/helm/pkg/chart/loader"
	"helm.sh/helm/pkg/cli"
	"helm.sh/helm/pkg/release"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"strings"
	"time"
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
}

func NewDeploy() *Deploy {
	uid, _ := shortid.Generate()
	log := ctrl.Log.WithName("helm-" + strings.ToLower(uid))

	var d Deploy
	d.Log = log
	return &d
}

// we need to save the charts to cache to avoid repeat download
var chartCache = cache.New(30*time.Minute, 60*time.Minute)

// Sync = install + upgrade
// When sync done, add the release note to HelmRequest status
// inCluster info is used to retrieve config info for valuesFrom
func (d *Deploy) Sync() (*release.Release, error) {
	log := d.Log
	hr := d.HelmRequest

	name := getReleaseName(hr)
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
	client.Namespace = hr.Spec.Namespace
	client.Install = true
	// This should be a reasonable value
	client.MaxHistory = 10

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

	downloader := NewDownloader(d.SystemNamespace, d.InCluster.ToRestConfig(), d.Log)
	chartPath, err := downloader.downloadChart(hr.Spec.Chart, hr.Spec.Version)
	if err != nil {
		return nil, err
	}
	log.Info("load charts from disk", "path", chartPath)
	ch, err = loader.Load(chartPath)
	if err != nil {
		return nil, err
	}

	if req := ch.Metadata.Dependencies; req != nil {
		if err := action.CheckDependencies(ch, req); err != nil {
			return nil, err
		}
	}

	if !d.Deployed {
		log.Info("Release does not exist. Installing it now", "name", name)
		// emptyValues := map[string]interface{}{}
		// rel := createRelease(cfg, ch, name, client.Namespace, emptyValues)
		resp, err := d.install()
		if err != nil {
			// if error occurred, just return. Otherwise the upgrade will stuck at no deploy found
			log.Error(err, "install before upgrade failed", "name", hr.Name)
			return resp, err
		}
		hr.Status.Notes = resp.Info.Notes
		return resp, nil
	}

	// run upgrade/install
	resp, err := client.Run(name, ch, values)
	if err != nil {
		return nil, errors.Wrap(err, "UPGRADE FAILED")
	}
	action.PrintRelease(out, resp)
	log.Info("Release has been upgraded. Happy Helming!\n", "name", name)

	// Print the status like status command does
	statusClient := action.NewStatus(cfg)
	rel, err := statusClient.Run(name)
	if err != nil {
		log.Error(err, "print status error")

	}
	action.PrintRelease(out, rel)
	if rel != nil {
		hr.Status.Notes = rel.Info.Notes
	}
	return resp, nil

}
