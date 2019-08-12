package webhook

import (
	"github.com/alauda/captain/pkg/apis/app/v1alpha1"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func validateChartRepoHandler(ws *webhook.Server) error {
	handler := admission.ValidatingWebhookFor(&v1alpha1.ChartRepo{})
	if err := handler.InjectLogger(log.Log.WithName("validate-chartrepo")); err != nil {
		klog.Error("inject logger to chartrepo validating webhook handler error: ", err)
		return err
	}
	ws.Register("/validate-chartrepo", handler)
	return nil
}

//RegisterHandlers register validating and mutating webhook for captain
func RegisterHandlers(mgr manager.Manager) error {
	// get and add it to manager
	ws := mgr.GetWebhookServer()

	if err := validateChartRepoHandler(ws); err != nil {
		return err
	}

	handler := admission.ValidatingWebhookFor(&v1alpha1.HelmRequest{})
	if err := handler.InjectLogger(log.Log.WithName("validating")); err != nil {
		klog.Error("inject logger to validating webhook handler error: ", err)
		return err
	}
	ws.Register("/validate", handler)

	handler = admission.DefaultingWebhookFor(&v1alpha1.HelmRequest{})
	if err := handler.InjectLogger(log.Log.WithName("mutating")); err != nil {
		klog.Error("inject logger to mutating webhook handler error: ", err)
		return err
	}
	ws.Register("/mutate", handler)
	return nil

}
