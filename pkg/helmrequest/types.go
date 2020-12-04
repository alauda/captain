package helmrequest

import (
	"github.com/alauda/helm-crds/pkg/apis/app/v1alpha1"
)

var (
	// AutoResolveVersionOnce is intend to indicate captain that once it resolve a latest version for this HelmRequest,
	// in the following updates, skip auto-resolve the version again.
	AutoResolveVersionOnce = "auto-resolve-version-once"
)

// ResolveVersion resolve helmrequest version for install.
// 1. If hr.spec.version is not empty, use it
// 2. If hr.spec.version is empty
//    a. If annotations.auto-resolve-version-once=false, auto resolve version every time from chartrepo
//    b. If annotations.auto-resolve-version-once=true, auto resolve version in the first time, then use the version from
//       .status.verion
func ResolveVersion(hr *v1alpha1.HelmRequest) string {
	version := hr.Spec.Version
	if version != "" {
		return version
	}
	if hr.Annotations == nil || hr.Annotations[AutoResolveVersionOnce] == "false" {
		return version
	}
	return hr.Status.Version
}
