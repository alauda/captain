package controller

import (
	"github.com/alauda/captain/pkg/apis/app/v1alpha1"
	"github.com/alauda/captain/pkg/helm"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

// updateChartRepoStatus update ChartRepo's status
func (c *Controller) updateChartRepoStatus(cr *v1alpha1.ChartRepo, phase v1alpha1.ChartRepoPhase, reason string) {
	cr = cr.DeepCopy()
	cr.Status.Phase = phase
	cr.Status.Reason = reason

	_, err := c.hrClientSet.AppV1alpha1().ChartRepos(cr.Namespace).Update(cr)
	if err != nil {
		klog.Error("update chartrepo error: ", err)
	}
}

// syncChartRepo sync ChartRepo to helm repo store
func (c *Controller) syncChartRepo(obj interface{}) {

	cr := obj.(*v1alpha1.ChartRepo)

	var username string
	var password string

	if cr.Spec.Secret != nil {
		ns := cr.Spec.Secret.Namespace
		if ns == "" {
			ns = cr.Namespace
		}
		secret, err := c.kubeClient.CoreV1().Secrets(ns).Get(cr.Spec.Secret.Name, v1.GetOptions{})
		if err != nil {
			c.updateChartRepoStatus(cr, v1alpha1.ChartRepoFailed, err.Error())
			klog.Error("get secret for chartrepo error: ", err)
			return
		}
		data := secret.Data
		username = string(data["username"])
		password = string(data["password"])

	}

	if err := helm.AddBasicAuthRepository(cr.GetName(), cr.Spec.URL, username, password); err != nil {
		c.updateChartRepoStatus(cr, v1alpha1.ChartRepoFailed, err.Error())
		return
	}

	c.updateChartRepoStatus(cr, v1alpha1.ChartRepoSynced, "")
	klog.Info("synced chartrepo: ", cr.GetName())
	return

}
