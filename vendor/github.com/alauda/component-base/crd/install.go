package crd

import (
	"reflect"

	extensionsobj "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
)

// EnsureCRDCreated tries to create/update crd, returns (true, nil) if succeeding, otherwise returns (false, nil).
// 'err' should always be nil, because it is used by wait.PollUntil(), and it will exit if it is not nil.
func EnsureCRDCreated(config *rest.Config, crd *extensionsobj.CustomResourceDefinition) (created bool, err error) {
	client, err := apiextensionsclient.NewForConfig(config)
	if err != nil {
		return false, err
	}
	return ensureCRDCreated(client, crd)
}

// ensureCRDCreated create or update a crd
// This function is so common, may be should move it to a library
func ensureCRDCreated(client apiextensionsclient.Interface, crd *extensionsobj.CustomResourceDefinition) (created bool, err error) {
	crdClient := client.ApiextensionsV1beta1().CustomResourceDefinitions()
	presetCRD, err := crdClient.Get(crd.Name, metav1.GetOptions{})
	if err == nil {
		if reflect.DeepEqual(presetCRD.Spec, crd.Spec) {
			klog.V(1).Infof("crd %s already exists", crd.Name)
		} else {
			klog.V(3).Infof("Update crd %s: %+v -> %+v", crd.Name, presetCRD.Spec, crd.Spec)
			newCRD := crd
			newCRD.ResourceVersion = presetCRD.ResourceVersion
			// Update crd
			if _, err := crdClient.Update(newCRD); err != nil {
				klog.Errorf("Error update crd %s: %v", crd.Name, err)
				return false, nil
			}
			klog.V(1).Infof("Update crd %s successfully.", crd.Name)
		}
	} else {
		// If not exist, create a new one
		// ????? why there is resourceVersion
		klog.Infof("creating crd: %+v", crd)
		crd.ResourceVersion = ""
		if _, err := crdClient.Create(crd); err != nil {
			klog.Errorf("Error creating crd %s: %v", crd.Name, err)
			return false, nil
		}
		klog.V(1).Infof("Create crd %s successfully.", crd.Name)
	}

	return true, nil
}
