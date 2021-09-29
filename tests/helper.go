package tests

import (
	"io/ioutil"
	"strings"

	"github.com/alauda/helm-crds/pkg/apis/app/v1beta1"
	"helm.sh/helm/v3/pkg/repo"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func ReadFile(path string) string {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	return string(data)
}

// LoadChart ...
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

// LoadIndex load index from file
func LoadIndex(path string) *repo.IndexFile {
	s := ReadFile(path)
	sr := strings.NewReader(s)
	d := yaml.NewYAMLOrJSONDecoder(sr, len(s))
	index := &repo.IndexFile{}
	if err := d.Decode(index); err != nil {
		panic(err)
	}
	return index
}
