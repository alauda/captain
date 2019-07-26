package webhook

import (
	"github.com/alauda/captain/pkg/apis/app/v1alpha1"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

//RegisterHandlers register validating and mutating webhook for captain
func RegisterHandlers(mgr manager.Manager) error {
	// get and add it to manager
	ws := mgr.GetWebhookServer()

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
