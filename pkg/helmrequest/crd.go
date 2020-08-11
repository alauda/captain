package helmrequest

import (
	"github.com/alauda/component-base/crd"
	"github.com/alauda/helm-crds/pkg/apis/app/v1alpha1"
	extensionsobj "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

// CRD is the CRD definition for HelmRequest
// CRD file and example is in $CAPTAIN_ROOT/artifacts/crd/*.yaml
var CRD = &extensionsobj.CustomResourceDefinition{
	ObjectMeta: metav1.ObjectMeta{
		Name: "helmrequests." + v1alpha1.SchemeGroupVersion.Group,
	},
	TypeMeta: metav1.TypeMeta{
		Kind:       "CustomResourceDefinition",
		APIVersion: "apiextensions.k8s.io/v1beta1",
	},
	Spec: extensionsobj.CustomResourceDefinitionSpec{
		Group:   v1alpha1.SchemeGroupVersion.Group,
		Version: v1alpha1.SchemeGroupVersion.Version,
		Scope:   extensionsobj.ResourceScope("Namespaced"),
		Names: extensionsobj.CustomResourceDefinitionNames{
			Plural:     "helmrequests",
			Singular:   "helmrequest",
			Kind:       "HelmRequest",
			ListKind:   "HelmRequestList",
			ShortNames: []string{"hr"},
			Categories: []string{"all"},
		},
		AdditionalPrinterColumns: []extensionsobj.CustomResourceColumnDefinition{
			{
				Name:     "Chart",
				Type:     "string",
				JSONPath: ".spec.chart",
			},
			{
				Name:     "Version",
				Type:     "string",
				JSONPath: ".spec.version",
			},
			{
				Name:     "Namespace",
				Type:     "string",
				JSONPath: ".spec.namespace",
			},
			{
				Name:     "AllCluster",
				Type:     "boolean",
				JSONPath: ".spec.installToAllClusters",
			},
			{
				Name:     "Age",
				Type:     "date",
				JSONPath: ".metadata.creationTimestamp",
			},
			{
				Name:     "Phase",
				Type:     "string",
				JSONPath: ".status.phase",
			},
		},
		Subresources: &extensionsobj.CustomResourceSubresources{
			Status: &extensionsobj.CustomResourceSubresourceStatus{},
		},
		Validation: &extensionsobj.CustomResourceValidation{
			OpenAPIV3Schema: &extensionsobj.JSONSchemaProps{
				Properties: map[string]extensionsobj.JSONSchemaProps{
					"spec": {
						Required: []string{
							"chart",
						},
						Properties: map[string]extensionsobj.JSONSchemaProps{
							"values": {
								Type: "object",
							},
						},
					},
				},
			},
		},
	},
}

// EnsureCRDCreated tries to create/update CRD, returns (true, nil) if succeeding, otherwise returns (false, nil).
// 'err' should always be nil, because it is used by wait.PollUntil(), and it will exit if it is not nil.
func EnsureCRDCreated(cfg *rest.Config) (created bool, err error) {
	return crd.EnsureCRDCreated(cfg, CRD)
	// return util.EnsureCRDCreated(cfg, CRD)
}
