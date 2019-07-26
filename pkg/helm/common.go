package helm

import (
	"alauda.io/captain/pkg/apis/app/v1alpha1"
	"github.com/alauda/component-base/hash"
)

//IsHelmRequestSynced check if a HelmRequest is synced
// only if hash is equal and not install to all clusters
func IsHelmRequestSynced(hr *v1alpha1.HelmRequest) bool {
	current := hash.GenHashStr(hr.Spec)
	return current == hr.Status.LastSpecHash
}

// getReleaseName get release name
func getReleaseName(hr *v1alpha1.HelmRequest) string {
	name := hr.GetName()
	if hr.Spec.ReleaseName != "" {
		name = hr.Spec.ReleaseName
	}
	return name
}
