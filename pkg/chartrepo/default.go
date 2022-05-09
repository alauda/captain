package chartrepo

import (
	"context"
	"time"

	"github.com/alauda/helm-crds/pkg/apis/app/v1beta1"
	clientset "github.com/alauda/helm-crds/pkg/client/clientset/versioned"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
)

func InstallDefaultChartRepo(cfg *rest.Config, ns string) error {
	client, err := clientset.NewForConfig(cfg)
	if err != nil {
		return err
	}

	return wait.PollImmediateUntil(time.Second*5, func() (bool, error) {
		return installStableRepo(client, ns)
	}, context.TODO().Done())
}

func installStableRepo(client *clientset.Clientset, ns string) (bool, error) {
	data := v1beta1.ChartRepo{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "stable",
			Namespace: ns,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "ChartRepo",
			APIVersion: "app.alauda.io/v1beta1",
		},
		Spec: v1beta1.ChartRepoSpec{
			URL: "https://kubernetes-charts.storage.googleapis.com",
		},
	}

	_, err := client.AppV1beta1().ChartRepos(ns).Create(&data)
	if err != nil {
		if errors.IsAlreadyExists(err) {
			klog.Info("default chartrepo already exist, skip...")
			return true, nil
		}
		return false, err
	}
	return true, nil
}
