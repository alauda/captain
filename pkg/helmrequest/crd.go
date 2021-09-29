package helmrequest

import (
	"context"
	"reflect"
	"strings"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
)

// CreateCRDObject create a crd object from yaml string
func CreateCRDObject(s string) (*apiextensionsv1.CustomResourceDefinition, error) {
	sr := strings.NewReader(s)
	d := yaml.NewYAMLOrJSONDecoder(sr, len(s))
	crdVar := &apiextensionsv1.CustomResourceDefinition{}
	if err := d.Decode(crdVar); err != nil {
		return nil, err
	}
	return crdVar, nil

}

// EnsureCRDCreated tries to create/update CRD, returns (true, nil) if succeeding, otherwise returns (false, nil).
// 'err' should always be nil, because it is used by wait.PollUntil(), and it will exit if it is not nil.
func EnsureCRDCreated(cfg *rest.Config) (created bool, err error) {
	crdVar, err := CreateCRDObject(helmRequestCRDYaml)
	if err != nil {
		return false, err
	}
	return EnsureCRDCreatedWithConfig(cfg, crdVar)
}

// EnsureCRDCreatedWithConfig tries to create/update crd, returns (true, nil) if succeeding, otherwise returns (false, nil).
// 'err' should always be nil, because it is used by wait.PollUntil(), and it will exit if it is not nil.
func EnsureCRDCreatedWithConfig(config *rest.Config, crd *apiextensionsv1.CustomResourceDefinition) (created bool, err error) {
	client, err := apiextensionsclient.NewForConfig(config)
	if err != nil {
		return false, err
	}
	return ensureCRDCreatedWithClient(client, crd)
}

// ensureCRDCreatedWithClient create or update a crd
// This function is so common, may be should move it to a library
func ensureCRDCreatedWithClient(client apiextensionsclient.Interface, crd *apiextensionsv1.CustomResourceDefinition) (created bool, err error) {
	crdClient := client.ApiextensionsV1().CustomResourceDefinitions()
	presetCRD, err := crdClient.Get(context.Background(), crd.Name, metav1.GetOptions{})
	if err == nil {
		if reflect.DeepEqual(presetCRD.Spec, crd.Spec) {
			klog.V(3).Infof("crd %s already exists", crd.Name)
		} else {
			klog.V(3).Infof("Update crd %s: %+v -> %+v", crd.Name, presetCRD.Spec, crd.Spec)
			newCRD := crd
			newCRD.ResourceVersion = presetCRD.ResourceVersion
			// Update crd
			if _, err := crdClient.Update(context.Background(), newCRD, metav1.UpdateOptions{}); err != nil {
				klog.Errorf("Error update crd %s: %v", crd.Name, err)
				return false, nil
			}
			klog.V(1).Infof("Update crd %s successfully.", crd.Name)
		}
	} else {
		// If not exist, create a new one
		// ????? why there is resourceVersion
		klog.V(3).Infof("creating crd: %s", crd.Name)
		crd.ResourceVersion = ""
		if _, err := crdClient.Create(context.Background(), crd, metav1.CreateOptions{}); err != nil {
			klog.Errorf("Error creating crd %s: %v", crd.Name, err)
			return false, nil
		}
		klog.V(3).Infof("Create crd %s successfully.", crd.Name)
	}

	return true, nil
}
