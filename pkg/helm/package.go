package helm

import (
	"errors"
	"fmt"
	"helm.sh/helm/pkg/action"
	"helm.sh/helm/pkg/chart/loader"
	"helm.sh/helm/pkg/chartutil"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	ctrl "sigs.k8s.io/controller-runtime"
	"strings"
)

var tempChartsDir = "/tmp/vcs-charts/"

var log = ctrl.Log.WithName("helm")

// uploadCharts upload charts from local dir to chartmuseum, repoName is the repo name
// as well as local sub-dir name
func uploadCharts(repoName string) error {
	url := fmt.Sprintf("%s/api/%s/charts", "http://captain-chartmuseum:8080", repoName)
	dir := tempChartsDir + repoName

	// if dir not exist, but no error found before, consider this VCS does not
	// contains valid helm charts. So we ignore this step, chartmusuem will provide
	// a default index
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		log.Info("vcs charts dir not exist, consider no charts found, skip upload", "name", repoName)
		return nil
	}

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Error(err, "read local charts dir error")
		return err
	}

	for _, f := range files {
		if !f.IsDir() {
			if err := postChart(url, filepath.Join(dir, f.Name())); err != nil {
				log.Error(err, "upload chart error", "filename", f.Name())
				return err
			}
			log.Info("upload chart done", "filename", f.Name())
		}
	}
	return nil

}

// SouceToChartRepo load vcs source to charts and upload them to chart repo
func SouceToChartRepo(name, dir, path string) error {
	if err := createCharts(name, dir, path); err != nil {
		return err
	}
	if err := uploadCharts(name); err != nil {
		return err
	}
	return nil

}

// createCharts create chart from source dir. We should support two scenario:
// 1. the path is a helm chart dir
// 2. the patch contains multiple chart
func createCharts(vcs, dir, path string) error {
	target := dir + "/" + path
	if path == "/" {
		target = dir
	}
	if strings.HasPrefix(path, "/") {
		target = dir + path
	}

	log.Info("vcs source dir is", "dir", target)

	files, err := ioutil.ReadDir(target)
	if err != nil {
		log.Error(err, "read vcs source dir error")
		return err
	}

	dest := tempChartsDir + vcs

	p := action.Package{
		Destination: dest,
	}

TOP:
	for _, f := range files {
		if !f.IsDir() && f.Name() == "Chart.yaml" {
			log.Info("vsc source dir is a helm package", "dir", target)
			_, err := p.Run(target, nil)
			if err != nil {
				log.Error(err, "package chart error")
				return err
			}
			return nil
		}

		if f.Name() == ".git" || f.Name() == ".svn" {
			continue
		}

		if f.IsDir() {
			sp := filepath.Join(target, f.Name())
			files, err := ioutil.ReadDir(sp)
			if err != nil {
				log.Error(err, "read vcs source sub dir error", "name", sp)
				return err
			}

			for _, f := range files {
				if !f.IsDir() && f.Name() == "Chart.yaml" {
					result, err := packageChart(sp, dest)
					if err != nil {
						log.Error(err, "package chart error")
						return err
					}
					log.Info("package chart done", "path", sp, "result", result)
					continue TOP
				}
			}

		}
	}
	return nil
}

// postChart post chart to repo
func postChart(url string, filename string) error {
	client := &http.Client{}
	data, err := os.Open(filename)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, data)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 300 {
		return errors.New("upload chart error, http status code is: " + resp.Status)
	}
	return nil
}

// a simple version of the helm package impl.
func packageChart(path, target string) (string, error) {
	ch, err := loader.LoadDir(path)
	if err != nil {
		return "", err
	}
	log.Info("load chart from local dir", "dir", path, "chart", ch.Metadata.Name)
	return chartutil.Save(ch, target)
}
