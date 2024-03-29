package chartrepo

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	clientset "github.com/alauda/helm-crds/pkg/client/clientset/versioned"
	"helm.sh/helm/v3/pkg/repo"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
)

func GetChartRepo(name string, ns string, cfg *rest.Config) (*repo.Entry, error) {
	client, err := clientset.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	coreClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	cr, err := client.AppV1beta1().ChartRepos(ns).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	var username string
	var password string

	if cr.Spec.Secret != nil {
		ns := cr.Spec.Secret.Namespace
		if ns == "" {
			ns = cr.Namespace
		}
		secret, err := coreClient.CoreV1().Secrets(ns).Get(context.Background(), cr.Spec.Secret.Name, metav1.GetOptions{})
		if err != nil {
			klog.Warningf("Get secret %s error for chartrepo %s", cr.Spec.Secret.Name, cr.Name)
			return nil, err
		}

		data := secret.Data
		username = string(data["username"])
		password = string(data["password"])
	}

	var entry repo.Entry
	entry.URL = cr.Spec.URL
	entry.Username = username
	entry.Password = password

	// parse url
	u, err := url.Parse(entry.URL)
	if err == nil {
		if u.User != nil {
			klog.Info("trying to separate username and password from url")
			entry.Username = u.User.Username()
			p, _ := u.User.Password()
			entry.Password = p
			entry.URL = fmt.Sprintf("%s://%s%s", u.Scheme, u.Host, u.Path)
			klog.Info("new url is: ", entry.URL)
		}
	}

	return &entry, nil

}

// GetChart get chart info, url and digest is the info we want
func GetChart(name, version, ns string, cfg *rest.Config) (*repo.ChartVersion, error) {
	client, err := clientset.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	chart, err := client.AppV1alpha1().Charts(ns).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	for _, item := range chart.Spec.Versions {
		if version == "" || version == item.Version {
			return &item.ChartVersion, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("cannot find version %s for chart %s", version, name))

}
