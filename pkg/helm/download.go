package helm

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/alauda/captain/pkg/chartrepo"
	"github.com/alauda/captain/pkg/registry"
	appv1 "github.com/alauda/helm-crds/pkg/apis/app/v1"
	"github.com/containers/image/v5/docker"
	"github.com/containers/image/v5/docker/reference"
	"github.com/go-logr/logr"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/repo"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	ChartsDir = "/tmp/helm-charts"

	transCfg = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // ignore expired SSL certificates
	}
	httpClient = &http.Client{Timeout: 30 * time.Second, Transport: transCfg}

	repoCache = cache.New(5*time.Minute, 10*time.Minute)
)

type Downloader struct {
	incfg *rest.Config
	cfg   *rest.Config

	// system ns
	ns string

	log logr.Logger
}

func NewDownloader(ns string, incfg, cfg *rest.Config, log logr.Logger) *Downloader {
	return &Downloader{
		incfg: incfg,
		cfg:   cfg,
		ns:    ns,
		log:   log,
	}
}

// stable/nginx -> stable nginx
func getRepoAndChart(name string) (string, string) {
	data := strings.Split(name, "/")
	if len(data) != 2 {
		return "", ""
	}
	return data[0], data[1]
}

// get from cache first, then get from k8s
func (d *Downloader) getRepoInfo(name string, ns string) (*repo.Entry, error) {
	result, ok := repoCache.Get(name)
	if ok {
		return result.(*repo.Entry), nil
	}
	entry, err := chartrepo.GetChartRepo(name, ns, d.incfg)
	if err == nil {
		repoCache.SetDefault(name, entry)
	}
	return entry, err
}

// downloadChart download a chart from helm repo to local disk and return the chart
// name: <repo>/<chart>
func (d *Downloader) downloadChart(name string, version string) (*chart.Chart, error) {
	log := d.log

	repoName, chart := getRepoAndChart(name)
	if repoName == "" && chart == "" {
		return nil, errors.New("cannot parse chart name")
	}
	log.Info("get chart", "name", name, "version", version)

	dir := ChartsDir
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}
		log.Info("helm charts dir not exist, create it: ", "dir", dir)
	}

	entry, err := d.getRepoInfo(repoName, d.ns)
	if err != nil {
		log.Error(err, "get chartrepo error")
		return nil, err
	}

	chartResourceName := fmt.Sprintf("%s.%s", strings.ToLower(chart), repoName)

	cv, err := chartrepo.GetChart(chartResourceName, version, d.ns, d.incfg)
	if err != nil {
		log.Error(err, "get chart error")
		return nil, err
	}

	path := cv.URLs[0]

	fileName := splitChartNameFromURL(path)
	filePath := fmt.Sprintf("%s/%s-%s-%s", dir, repoName, cv.Digest, fileName)

	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		log.Info("chart already downloaded, use it", "path", filePath)
		return loader.Load(filePath)
	}

	return loadFileFromEntry(entry, path, filePath)
}

// loadFileFromEntry will download a url and store it in local filepath.
// It writes to the destination file as it downloads it, without
// loading the entire file into memory.
func loadFileFromEntry(entry *repo.Entry, chartPath, filePath string) (*chart.Chart, error) {
	ep := entry.URL + "/" + chartPath
	if strings.HasSuffix(entry.URL, "/") {
		ep = entry.URL + chartPath
	}

	if strings.HasPrefix(chartPath, "http://") || strings.HasPrefix(chartPath, "https://") {
		ep = chartPath
	}

	return loadChart(ep, entry.Username, entry.Password, filePath)
}

func loadChart(url, username, password, filePath string) (*chart.Chart, error) {
	req, err := http.NewRequest("GET", url, nil)
	if username != "" && password != "" {
		req.SetBasicAuth(username, password)
	}

	// Get the data
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.Errorf("failed to fetch %s : %s", url, resp.Status)
	}

	buf := bytes.NewBuffer(nil)
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		return nil, err
	}

	if err := ioutil.WriteFile(filePath, buf.Bytes(), 0644); err != nil {
		return nil, err
	}

	return loader.Load(filePath)
}

func (d *Downloader) downloadChartFromHTTP(hr *appv1.HelmRequest) (*chart.Chart, error) {
	var err error
	if hr.Spec.Source != nil && hr.Spec.Source.HTTP != nil {
		if hr.Spec.Source.HTTP.URL != "" {
			dir := ChartsDir
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				if err = os.MkdirAll(dir, 0755); err != nil {
					return nil, err
				}
				log.Info("helm charts dir not exist, create it: ", "dir", dir)
			}

			url := hr.Spec.Source.HTTP.URL
			chname := splitChartNameFromURL(url)

			filePath := fmt.Sprintf("%s/%s-%s", dir, hr.GetName(), chname)
			if _, err := os.Stat(filePath); !os.IsNotExist(err) {
				log.Info("chart already downloaded, remove it", "path", filePath)
				os.Remove(filePath)
			}

			username, password := "", ""
			if hr.Spec.Source.HTTP.SecretRef != "" {
				username, password, err = d.fetchAuthFromSecret(hr.Spec.Source.HTTP.SecretRef, hr.GetNamespace())
				if err != nil {
					return nil, err
				}
			}

			if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
				return loadChart(url, username, password, filePath)
			}
			err = errors.New("helmrequest spec source http url does not start with HTTP or HTTPS")
		} else {
			err = errors.New("helmrequest spec source http url not found")
		}
	}

	if err == nil {
		err = errors.New("helmrequest spec source invalid, require HTTP type")
	}

	return nil, err
}

func (d *Downloader) pullOCIChart(hr *appv1.HelmRequest) (*chart.Chart, error) {
	if hr.Spec.Source != nil && hr.Spec.Source.OCI != nil {
		username, password := "", ""
		var err error
		if hr.Spec.Source.OCI.SecretRef != "" {
			username, password, err = d.fetchAuthFromSecret(hr.Spec.Source.OCI.SecretRef, hr.GetNamespace())
			if err != nil {
				return nil, err
			}
		}

		url := hr.Spec.Source.OCI.Repo + ":" + hr.Spec.Version
		return d.pullAndLoadChart(url, username, password)
	}

	return nil, errors.New("invalid chart Source, need OCI type")
}

func (d *Downloader) fetchAuthFromSecret(name, namespace string) (string, string, error) {
	inkc, err := kubernetes.NewForConfig(d.incfg)
	if err != nil {
		log.Error(err, "init kubernetes client error")
		return "", "", err
	}

	s, err := inkc.CoreV1().Secrets(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			kc, errI := kubernetes.NewForConfig(d.cfg)
			if errI != nil {
				log.Error(errI, "init incluster kubernetes client error")
				return "", "", errI
			}
			s, err = kc.CoreV1().Secrets(namespace).Get(context.Background(), name, metav1.GetOptions{})
			if err != nil {
				return "", "", err
			}
		} else {
			return "", "", err
		}
	}
	username, password := "", ""

	u, ok := s.Data["username"]
	if ok {
		username = strings.Trim(string(u), "\n")
	}
	p, ok := s.Data["password"]
	if ok {
		password = strings.Trim(string(p), "\n")
	}

	if username == "" || password == "" {
		return "", "", errors.New(fmt.Sprintf("can not find username or password in the secret %s/%s", namespace, name))
	}

	return username, password, nil
}

func splitChartNameFromURL(url string) string {
	if len(url) == 0 {
		return ""
	}

	idx := strings.LastIndex(url, "/")
	if idx == -1 {
		return url
	}
	return url[idx+1:]
}

func (d *Downloader) pullAndLoadChart(url, username, password string) (*chart.Chart, error) {
	client, err := registry.NewClient(
		registry.ClientOptDebug(true),
		registry.ClientOptPlainHTTP(false),
	)
	if err != nil {
		return nil, err
	}

	ref, err := registry.ParseReference(url)
	if err != nil {
		return nil, err
	}

	domain := ""
	if username != "" && password != "" {
		tmpURL := url
		if !strings.HasPrefix(tmpURL, "//") {
			tmpURL = fmt.Sprintf("//%s", url)
		}
		imageRef, err := docker.ParseReference(tmpURL)
		if err != nil {
			d.log.Error(err, "could not parse image")
			return nil, err
		}
		domain = reference.Domain(imageRef.DockerReference())
		if err := client.Login(domain, username, password, true); err != nil {
			d.log.Error(err, "registry login error")
		}
	}

	buffer, err := client.PullChart(ref, true)
	if err != nil {
		if strings.Contains(err.Error(), "server gave HTTP response to HTTPS client") {
			d.log.Info("Will try to pull chart again with plainHTTP", "response", err.Error())
			// set plainHTTP to true, and try to pull again
			client2, err := registry.NewClient(
				registry.ClientOptDebug(true),
				registry.ClientOptPlainHTTP(true),
			)
			if err != nil {
				return nil, err
			}
			if domain != "" && username != "" && password != "" {
				if err := client2.Login(domain, username, password, true); err != nil {
					d.log.Error(err, "registry login error")
				}
			}

			buffer, err = client2.PullChart(ref, true)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	d.log.Info("Pull chart successfully", "url", url)
	return loader.LoadArchive(buffer)
}
