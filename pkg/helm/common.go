package helm

import (
	"fmt"
	"hash"
	"hash/fnv"

	"github.com/alauda/captain/pkg/util"
	appv1 "github.com/alauda/helm-crds/pkg/apis/app/v1"
	"github.com/davecgh/go-spew/spew"
	"helm.sh/helm/v3/pkg/helmpath"
)

type ChartSourceType string

const (
	ChartSourceHTTP  ChartSourceType = "http"
	ChartSourceOCI   ChartSourceType = "oci"
	ChartSourceChart ChartSourceType = "chart"
)

var systemUsers []string = []string{"admin", "kubernetes-admin"}

// DeepHashObject writes specified object to hash using the spew library
// which follows pointers and prints actual values of the nested objects
// ensuring the hash does not change when a pointer changes.
func DeepHashObject(hasher hash.Hash, objectToWrite interface{}) {
	// copy from k8s.io/kubernetes/pkg/util/hash/hash.go to avoid import this monster
	hasher.Reset()
	printer := spew.ConfigState{
		Indent:         " ",
		SortKeys:       true,
		DisableMethods: true,
		SpewKeys:       true,
	}
	printer.Fprintf(hasher, "%#v", objectToWrite)
}

func genHash(data interface{}) uint64 {
	hasher := fnv.New64()
	DeepHashObject(hasher, data)
	return hasher.Sum64()
}

//GenHashStr ... This is a very stupid function, need to learn from kubernetes at how to
// generate pod name for deployment
func GenHashStr(data interface{}) string {
	s := fmt.Sprintf("%d", genHash(data))
	return s
}

// GenUniqueHash generate a unique hash for a HelmRequest
func GenUniqueHash(hr *appv1.HelmRequest) string {
	source := struct {
		spec        appv1.HelmRequestSpec
		annotations map[string]string
	}{
		hr.Spec,
		hr.Annotations,
	}
	return GenHashStr(source)
}

// IsHelmRequestSynced check if a HelmRequest is synced
// only if hash is equal and not install to all clusters
// First version: only hash .spec
// Second version: hash .spec and .metadata.annotations
func IsHelmRequestSynced(hr *appv1.HelmRequest) bool {
	current := GenUniqueHash(hr)
	if current == hr.Status.LastSpecHash {
		return true
	}

	// This is for old and exist HelmRequest, they are already synced.
	// we don't want the algorithm change to cause them to be upgrade again
	onlySpec := GenHashStr(hr.Spec)
	if onlySpec == hr.Status.LastSpecHash {
		if hr.Annotations == nil || hr.Annotations[util.KubectlCaptainSync] == "" {
			return true
		}
	}

	return false
}

// GetReleaseName get release name
func GetReleaseName(hr *appv1.HelmRequest) string {
	return hr.GetReleaseName()
}

func helmRepositoryFile() string {
	return helmpath.ConfigPath("repositories.yaml")
}

// isSwitchEnabled return annoKey Annotation is true or not
func isSwitchEnabled(hr *appv1.HelmRequest, annoKey string) bool {
	if hr == nil || len(hr.Annotations) == 0 {
		return false
	}

	if hr.Annotations[annoKey] == "true" {
		return true
	}

	return false
}

func getChartSourceType(hr *appv1.HelmRequest) ChartSourceType {
	if hr != nil && hr.Spec.Source != nil {
		if hr.Spec.Source.HTTP != nil {
			return ChartSourceHTTP
		}

		if hr.Spec.Source.OCI != nil {
			return ChartSourceOCI
		}
	}

	return ChartSourceChart
}
