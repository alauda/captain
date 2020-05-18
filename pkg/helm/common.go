package helm

import (
	"github.com/alauda/captain/pkg/util"
	"github.com/alauda/component-base/hash"
	"github.com/alauda/helm-crds/pkg/apis/app/v1beta1"
	"helm.sh/helm/pkg/helmpath"
)

// GenUniqueHash generate a unique hash for a HelmRequest
func GenUniqueHash(hr *v1beta1.HelmRequest) string {
	source := struct {
		spec        v1beta1.HelmRequestSpec
		annotations map[string]string
	}{
		hr.Spec,
		hr.Annotations,
	}
	return hash.GenHashStr(source)
}

// IsHelmRequestSynced check if a HelmRequest is synced
// only if hash is equal and not install to all clusters
// First version: only hash .spec
// Second version: hash .spec and .metadata.annotations
func IsHelmRequestSynced(hr *v1beta1.HelmRequest) bool {
	current := GenUniqueHash(hr)
	if current == hr.Status.LastSpecHash {
		return true
	}

	// This is for old and exist HelmRequest, they are already synced.
	// we don't want the algorithm change to cause them to be upgrade again
	onlySpec := hash.GenHashStr(hr.Spec)
	if onlySpec == hr.Status.LastSpecHash {
		if hr.Annotations == nil || hr.Annotations[util.KubectlCaptainSync] == "" {
			return true
		}
	}

	return false
}

// getReleaseName get release name
func getReleaseName(hr *v1beta1.HelmRequest) string {
	name := hr.GetName()
	if hr.Spec.ReleaseName != "" {
		name = hr.Spec.ReleaseName
	}
	return name
}

func helmRepositoryFile() string {
	return helmpath.ConfigPath("repositories.yaml")
}
