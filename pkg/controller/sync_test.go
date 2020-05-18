package controller

import (
	"testing"

	"github.com/alauda/helm-crds/pkg/apis/app/v1beta1"

	"github.com/gsamokovarov/assert"

	"github.com/alauda/component-base/hash"
	"github.com/ghodss/yaml"
	"helm.sh/helm/pkg/chartutil"
)

func TestHelmRequestDeepCopyHash(t *testing.T) {
	values := `
    global:
      registry:
        address: 10.0.128.234:60080
    replicas: 1
    resources:
      requests:
        cpu: 10m
        memory: 10m`
	var v chartutil.Values
	yaml.Unmarshal([]byte(values), &v)
	hr := &v1beta1.HelmRequest{
		Spec: v1beta1.HelmRequestSpec{
			Chart:                "stable/captain-test-demo",
			InstallToAllClusters: true,
			Namespace:            "default",
			ReleaseName:          "cpatain-test-demo",
			HelmValues:           v1beta1.HelmValues{Values: v},
			Version:              "1.2.1",
		},
	}
	assert.Equal(t, hash.GenHashStr(hr.Spec), hash.GenHashStr(hr.DeepCopy().Spec))
}
