package webhook

import (
	"encoding/base64"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

var wLog = ctrl.Log.WithName("webhook")

func InjectCertToWebhook(data []byte, cfg *rest.Config) error {
	// cache not populated yet...
	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return err
	}

	decoded, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		return err
	}

	mw, err := client.AdmissionregistrationV1beta1().MutatingWebhookConfigurations().Get("captain-mutating-webhook-configuration", metav1.GetOptions{})
	if err != nil {
		return err
	}

	if string(mw.Webhooks[0].ClientConfig.CABundle) == "Cg==" {
		wLog.Info("mutating webhook contains no cert data, update now")

		mw.Webhooks[0].ClientConfig.CABundle = decoded
		if _, err := client.AdmissionregistrationV1beta1().MutatingWebhookConfigurations().Update(mw); err != nil {
			return err
		}

	}

	vw, err := client.AdmissionregistrationV1beta1().ValidatingWebhookConfigurations().Get("captain-validating-webhook-configuration", metav1.GetOptions{})
	if err != nil {
		return err

	}

	if string(vw.Webhooks[0].ClientConfig.CABundle) == "Cg==" {
		wLog.Info("validating webhook contains no cert data, update now")
		vw.Webhooks[0].ClientConfig.CABundle = decoded

		if _, err := client.AdmissionregistrationV1beta1().ValidatingWebhookConfigurations().Update(vw); err != nil {
			return err
		}

	}
	return nil

}
