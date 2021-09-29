package webhook

import (
	"context"
	"encoding/base64"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
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

	mw, err := client.AdmissionregistrationV1beta1().MutatingWebhookConfigurations().Get(context.Background(), "captain-mutating-webhook-configuration", metav1.GetOptions{})
	if err != nil {
		wLog.Error(err, "get mutate webhook error, ignore")
		return nil
	}

	// equal := bytes.Compare(mw.Webhooks[0].ClientConfig.CABundle, decoded)

	// wLog.Info("debug webhook data", "mw", mw)
	// wLog.Info("debug data", "ca", string(mw.Webhooks[0].ClientConfig.CABundle[:]), "equal", equal == 0)

	mw.Webhooks[0].ClientConfig.CABundle = decoded
	if _, err := client.AdmissionregistrationV1beta1().MutatingWebhookConfigurations().Update(context.Background(), mw, metav1.UpdateOptions{}); err != nil {
		return err
	}

	vw, err := client.AdmissionregistrationV1beta1().ValidatingWebhookConfigurations().Get(context.Background(), "captain-validating-webhook-configuration", metav1.GetOptions{})
	if err != nil {
		wLog.Error(err, "get valite webhook error , ignore")
		return nil
	}

	vw.Webhooks[0].ClientConfig.CABundle = decoded

	if _, err := client.AdmissionregistrationV1beta1().ValidatingWebhookConfigurations().Update(context.Background(), vw, metav1.UpdateOptions{}); err != nil {
		return err
	}

	return nil

}
