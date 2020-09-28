package tests

import (
	"github.com/alauda/helm-crds/pkg/apis/app/v1beta1"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/util/yaml"
	"strings"
)

func ReadFile(path string) string {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	return string(data)
}

// createCRDObject create a crd object from yaml string
func LoadChart(path string) *v1beta1.Chart {
	s := ReadFile(path)
	sr := strings.NewReader(s)
	d := yaml.NewYAMLOrJSONDecoder(sr, len(s))
	chart := &v1beta1.Chart{}
	if err := d.Decode(chart); err != nil {
		panic(err)
	}
	return chart
}
