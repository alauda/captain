package helm

import (
	"os"
	"time"

	"github.com/alauda/captain/pkg/util"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/release"
)

//install install a chart to a cluster, If the release already exist, upgrade it
func (d *Deploy) install(chart *chart.Chart) (*release.Release, error) {
	hr := d.HelmRequest
	inCluster := d.InCluster
	log := d.Log

	cfg, err := d.newActionConfig()
	if err != nil {
		return nil, err
	}
	client := action.NewInstall(cfg)
	// This is used for crd-install webhook, or it will wait forever
	client.Timeout = 180 * time.Second
	out := os.Stdout
	settings := cli.New()
	settings.Debug = true

	args := []string{
		hr.GetName(),
		hr.Spec.Chart,
	}
	_, chrt, err := client.NameAndChart(args)
	if err != nil {
		d.Log.Error(err, "get chart and name error")
		return nil, err
	}

	log.Info("chart name is", "name", chrt)

	client.ReleaseName = GetReleaseName(hr)
	client.Replace = true
	// Do not validate openapi schema
	client.DisableOpenAPIValidation = true
	client.ForceAdopt = isSwitchEnabled(hr, util.ForceAdoptResourcesAnnotation)

	if hr.Spec.Version != "" {
		client.ChartPathOptions = action.ChartPathOptions{
			Version: hr.Spec.Version,
		}
	}

	values, err := getValues(hr, inCluster.ToRestConfig())
	if err != nil {
		return nil, err
	}

	client.Namespace = hr.GetReleaseNamespace()
	validInstallableChart, err := isChartInstallable(chart)
	if !validInstallableChart {
		log.Error(err, "not installable error")
		return nil, err
	}
	d.HelmRequest.Status.Version = chart.Metadata.Version

	if req := chart.Metadata.Dependencies; req != nil {
		// If CheckDependencies returns an error, we have unfulfilled dependencies.
		// As of Helm 2.4.0, this is treated as a stopping condition:
		// https://github.com/helm/helm/issues/2209
		if err := action.CheckDependencies(chart, req); err != nil {
			if client.DependencyUpdate {
				man := &downloader.Manager{
					Out:        out,
					ChartPath:  ChartsDir,
					Keyring:    client.ChartPathOptions.Keyring,
					SkipUpdate: false,
					Getters:    getter.All(settings),
				}
				if err := man.Update(); err != nil {
					return nil, err
				}
			} else {
				return nil, err
			}
		}
	}

	return client.Run(chart, values)
}

// isChartInstallable validates if a chart can be installed
//
// Application chart type is only installable
func isChartInstallable(ch *chart.Chart) (bool, error) {
	switch ch.Metadata.Type {
	case "", "application":
		return true, nil
	}
	return false, errors.Errorf("%s charts are not installable", ch.Metadata.Type)
}
