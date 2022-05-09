package webhook

import (
	appv1 "github.com/alauda/helm-crds/pkg/apis/app/v1"
	"github.com/alauda/helm-crds/pkg/apis/app/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func validateChartRepoHandler(ws *webhook.Server) error {
	handler := admission.ValidatingWebhookFor(&v1alpha1.ChartRepo{})
	if err := handler.InjectLogger(log.Log.WithName("validate-chartrepo")); err != nil {
		wLog.Error(err, "inject logger to chartrepo validating webhook handler error: ")
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

	handler := admission.ValidatingWebhookFor(&appv1.HelmRequest{})
	if err := handler.InjectLogger(log.Log.WithName("validating")); err != nil {
		wLog.Error(err, "inject logger to validating webhook handler error: ")
		return err
	}
	ws.Register("/validate", handler)

	handler = admission.DefaultingWebhookFor(&appv1.HelmRequest{})
	if err := handler.InjectLogger(log.Log.WithName("mutating")); err != nil {
		wLog.Error(err, "inject logger to mutating webhook handler error: ")
		return err
	}
	ws.Register("/mutate", handler)
	return nil

}
