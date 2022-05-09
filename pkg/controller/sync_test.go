package controller

import (
	"testing"

	"github.com/alauda/captain/pkg/helm"
	appv1 "github.com/alauda/helm-crds/pkg/apis/app/v1"
	"github.com/ghodss/yaml"
	"github.com/gsamokovarov/assert"
	"helm.sh/helm/v3/pkg/chartutil"
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
	hr := &appv1.HelmRequest{
		Spec: appv1.HelmRequestSpec{
			Chart:                "stable/captain-test-demo",
			InstallToAllClusters: true,
			Namespace:            "default",
			ReleaseName:          "cpatain-test-demo",
			HelmValues:           appv1.HelmValues{Values: v},
			Version:              "1.2.1",
		},
	}
	assert.Equal(t, helm.GenHashStr(hr.Spec), helm.GenHashStr(hr.DeepCopy().Spec))
}
