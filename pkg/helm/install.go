package helm

import (
	"os"
	"strings"
	"time"

	"github.com/alauda/captain/pkg/cluster"
	"github.com/alauda/helm-crds/pkg/apis/app/v1alpha1"
	"github.com/pkg/errors"
	"helm.sh/helm/pkg/action"
	"helm.sh/helm/pkg/chart"
	"helm.sh/helm/pkg/chart/loader"
	"helm.sh/helm/pkg/cli"
	"helm.sh/helm/pkg/downloader"
	"helm.sh/helm/pkg/getter"
	"helm.sh/helm/pkg/release"
	"k8s.io/klog"
)

//install install a chart to a cluster, If the release already exist, upgrade it
func install(hr *v1alpha1.HelmRequest, info *cluster.Info, inCluster *cluster.Info) (*release.Release, error) {
	cfg, err := newActionConfig(info)
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
		klog.Errorf("get chrt and name error: %s", err.Error())
		return nil, err
	}

	klog.Infof("Chart: %s", chrt)

	client.ReleaseName = getReleaseName(hr)
	// when install failed and we want to retry
	client.Replace = true

	if hr.Spec.Version != "" {
		client.ChartPathOptions = action.ChartPathOptions{
			Version: hr.Spec.Version,
		}
	}

	cp, err := client.ChartPathOptions.LocateChart(chrt, settings)
	if err != nil {
		klog.Errorf("locate chart %s error: %s", cp, err.Error())
		// a simple string match
		if client.Version == "" && strings.Contains(err.Error(), " no chart version found for") {
			klog.Info("no normal version found, try using devel flag")
			client.Version = ">0.0.0-0"
			cp, err = client.ChartPathOptions.LocateChart(chrt, settings)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	klog.V(9).Infof("CHART PATH: %s\n", cp)

	values, err := getValues(hr, inCluster.ToRestConfig())
	if err != nil {
		return nil, err
	}

	// Check chart dependencies to make sure all are present in /charts
	chartRequested, err := loader.Load(cp)
	if err != nil {
		klog.Errorf("load error: %s", err.Error())
		return nil, err
	}

	client.Namespace = hr.Spec.Namespace
	klog.Infof("load chart request: %s client: %+v", chartRequested.Name(), client)
	validInstallableChart, err := isChartInstallable(chartRequested)
	if !validInstallableChart {
		klog.Errorf("not installable error : %+v", err)
		return nil, err
	}

	if req := chartRequested.Metadata.Dependencies; req != nil {
		// If CheckDependencies returns an error, we have unfulfilled dependencies.
		// As of Helm 2.4.0, this is treated as a stopping condition:
		// https://github.com/helm/helm/issues/2209
		if err := action.CheckDependencies(chartRequested, req); err != nil {
			if client.DependencyUpdate {
				man := &downloader.Manager{
					Out:        out,
					ChartPath:  cp,
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

	return client.Run(chartRequested, values)
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
