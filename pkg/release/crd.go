package release

import (
	"time"

	"github.com/alauda/component-base/crd"

	"alauda.io/captain/pkg/apis/app/v1alpha1"
	extensionsobj "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
)

//CRD is the CRD definition for Release
// CRD file and example is in $CAPTAIN_ROOT/release/*.yaml
// Doc:
//   1. https://kubernetes.io/docs/tasks/access-kubernetes-api/custom-resources/custom-resource-definition-versioning/
//   2.
// TODO: do we need a status sub-resource?
var CRD = &extensionsobj.CustomResourceDefinition{
	ObjectMeta: metav1.ObjectMeta{
		Name: "releases." + v1alpha1.SchemeGroupVersion.Group,
	},
	TypeMeta: metav1.TypeMeta{
		Kind:       "CustomResourceDefinition",
		APIVersion: "apiextensions.k8s.io/v1beta1",
	},
	Spec: extensionsobj.CustomResourceDefinitionSpec{
		Group:   v1alpha1.SchemeGroupVersion.Group,
		Version: v1alpha1.SchemeGroupVersion.Version,
		Versions: []extensionsobj.CustomResourceDefinitionVersion{
			{
				Name:    v1alpha1.SchemeGroupVersion.Version,
				Served:  true,
				Storage: true,
			},
		},
		Conversion: &extensionsobj.CustomResourceConversion{
			Strategy: extensionsobj.NoneConverter,
		},
		Scope: extensionsobj.ResourceScope("Namespaced"),
		Names: extensionsobj.CustomResourceDefinitionNames{
			Plural:     "releases",
			Singular:   "release",
			Kind:       "Release",
			ListKind:   "ReleaseList",
			ShortNames: []string{"rel"},
			Categories: []string{"all"},
		},
		AdditionalPrinterColumns: []extensionsobj.CustomResourceColumnDefinition{
			{
				Name:     "Status",
				Type:     "string",
				JSONPath: ".status.status",
			},
			{
				Name:     "Age",
				Type:     "date",
				JSONPath: ".metadata.creationTimestamp",
			},
		},
	},
}

// EnsureCRDCreated tries to create/update CRD, returns (true, nil) if succeeding, otherwise returns (false, nil).
// 'err' should always be nil, because it is used by wait.PollUntil(), and it will exit if it is not nil.
func EnsureCRDCreated(cfg *rest.Config) error {
	return wait.PollImmediate(time.Second*3, time.Second*30, func() (bool, error) {
		return crd.EnsureCRDCreated(cfg, CRD)
	})
}
