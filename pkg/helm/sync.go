package helm

import (
	"os"

	"github.com/alauda/captain/pkg/apis/app/v1alpha1"
	"github.com/alauda/captain/pkg/cluster"
	"github.com/pkg/errors"
	"helm.sh/helm/pkg/action"
	"helm.sh/helm/pkg/chart/loader"
	"helm.sh/helm/pkg/cli"
	"helm.sh/helm/pkg/release"
	"helm.sh/helm/pkg/storage/driver"
	"k8s.io/klog"
)

// Sync = install + upgrade
// When sync done, add the release note to HelmRequest status
func Sync(hr *v1alpha1.HelmRequest, info *cluster.Info) (*release.Release, error) {
	name := getReleaseName(hr)
	out := os.Stdout

	// helm settings
	settings := cli.EnvSettings{
		Home:  getHelmHome(),
		Debug: true,
	}

	// init upgrade client
	cfg, err := newActionConfig(info)
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
	values, err := getValues(hr, info.ToRestConfig())
	if err != nil {
		return nil, err
	}
	client.ValueOptions = action.NewValueOptions(values)
	client.ResetValues = true

	// locate chart
	chrt := hr.Spec.Chart
	chartPath, err := client.ChartPathOptions.LocateChart(chrt, settings)
	if err != nil {
		klog.Errorf("locate chart %s error: %s", chartPath, err.Error())
		return nil, err
	}

	// load
	ch, err := loader.Load(chartPath)
	if err != nil {
		return nil, err
	}
	if req := ch.Metadata.Dependencies; req != nil {
		if err := action.CheckDependencies(ch, req); err != nil {
			return nil, err
		}
	}

	// since we set install to true, do a install first if not exist
	histClient := action.NewHistory(cfg)
	// big enough to contains all the history
	histClient.Max = 10000000
	histClient.OutputFormat = "json"
	if result, err := histClient.Run(name); err == driver.ErrReleaseNotFound || !isHaveDeployedRelease(result) {
		klog.Warningf("Release %q does not exist. Installing it now.\n", name)
		// emptyValues := map[string]interface{}{}
		// rel := createRelease(cfg, ch, name, client.Namespace, emptyValues)
		resp, err := install(hr, info)
		if err != nil {
			// if error occurred, just return. Otherwise the upgrade will stuck at not deploy found
			klog.Warning("install before upgrade failed: ", err)
			return resp, err
		}
		hr.Status.Notes = resp.Info.Notes
		return resp, nil
	}

	// run upgrade/install
	resp, err := client.Run(name, ch)
	if err != nil {
		return nil, errors.Wrap(err, "UPGRADE FAILED")
	}
	action.PrintRelease(out, resp)
	klog.Infof("Release %q has been upgraded. Happy Helming!\n", name)

	// Print the status like status command does
	statusClient := action.NewStatus(cfg)
	rel, err := statusClient.Run(name)
	if err != nil {
		klog.Warningf("print status error: %s", err.Error())
	}
	action.PrintRelease(out, rel)
	if rel != nil {
		hr.Status.Notes = rel.Info.Notes
	}
	return resp, nil

}

// isHaveDeployedRelease will check the history data to find out is there a successfully deployed release
func isHaveDeployedRelease(hist []*release.Release) bool {
	for _, item := range hist {
		if item.Info.Status == "deployed" {
			return true
		}
	}
	return false

}
