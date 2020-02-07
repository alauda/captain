/*
Copyright The Helm Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package storagedriver

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"k8s.io/klog"

	"github.com/alauda/helm-crds/pkg/apis/app/v1alpha1"
	"helm.sh/helm/pkg/chart"
	rspb "helm.sh/helm/pkg/release"
)

var b64 = base64.StdEncoding

var magicGzip = []byte{0x1f, 0x8b, 0x08}

// encodeRelease encodes a release returning a base64 encoded
// gzipped binary protobuf encoding representation, or error.
func encodeRelease(rls *rspb.Release) (*v1alpha1.Release, error) {
	var rel v1alpha1.Release

	data, err := encodeData(rls.Chart)
	if err != nil {
		return nil, err
	}
	rel.Spec.ChartData = data

	data, err = encodeData(rls.Config)
	if err != nil {
		return nil, err
	}
	rel.Spec.ConfigData = data

	data, err = encodeData(rls.Hooks)
	if err != nil {
		return nil, err
	}

	rel.Spec.HooksData = data

	data, err = encodeData(rls.Manifest)
	if err != nil {
		return nil, err
	}
	rel.Spec.ManifestData = data
	rel.Spec.Name = rls.Name
	rel.Spec.Version = rls.Version

	rel.Name = makeKey(rls.Name, rls.Version)
	rel.Namespace = rls.Namespace

	rel.Status.CopyFromReleaseInfo(rls.Info)
	return &rel, nil

}

func encodeData(data interface{}) (string, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	w, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		return "", err
	}
	if _, err = w.Write(b); err != nil {
		return "", err
	}
	if err := w.Close(); err != nil {
		return "", err
	}

	return b64.EncodeToString(buf.Bytes()), nil
}

func decodeRawRelease(rel *v1alpha1.Release) (*rspb.Release, error) {
	var rls rspb.Release
	rls.Info = rel.Status.ToReleaseInfo()
	rls.Version = rel.Spec.Version
	rls.Name = rel.Spec.Name
	rls.Namespace = rel.GetNamespace()
	return &rls, nil
}

// decodeRelease decodes the bytes in data into a release
// type. Data must contain a base64 encoded string of a
// valid protobuf encoding of a release, otherwise
// an error is returned.
func decodeRelease(rel *v1alpha1.Release) (*rspb.Release, error) {
	var rls rspb.Release
	rls.Info = rel.Status.ToReleaseInfo()

	rls.Version = rel.Spec.Version
	rls.Name = rel.Spec.Name
	rls.Namespace = rel.GetNamespace()

	var chart chart.Chart
	b, err := getEncodedBytes(rel.Spec.ChartData)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(b, &chart); err != nil {
		return nil, err
	}
	rls.Chart = &chart

	var config map[string]interface{}
	b, err = getEncodedBytes(rel.Spec.ConfigData)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(b, &config); err != nil {
		return nil, err
	}
	rls.Config = config

	var hooks []*rspb.Hook
	b, err = getEncodedBytes(rel.Spec.HooksData)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(b, &hooks); err != nil {
		klog.Errorf("decode hooks data error: %s", string(b))
		return nil, err
	}
	rls.Hooks = hooks

	var manifest string
	b, err = getEncodedBytes(rel.Spec.ManifestData)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(b, &manifest); err != nil {
		return nil, err
	}
	rls.Manifest = manifest

	return &rls, nil
}

func getEncodedBytes(data string) ([]byte, error) {
	// base64 decode string
	b, err := b64.DecodeString(data)
	if err != nil {
		return nil, err
	}

	// For backwards compatibility with releases that were stored before
	// compression was introduced we skip decompression if the
	// gzip magic header is not found
	if bytes.Equal(b[0:3], magicGzip) {
		r, err := gzip.NewReader(bytes.NewReader(b))
		if err != nil {
			return nil, err
		}
		b2, err := ioutil.ReadAll(r)
		if err != nil {
			return nil, err
		}
		b = b2
	}
	return b, nil
}

// makeKey concatenates a release name and version into
// a string with format ```<release_name>#v<version>```.
// This key is used to uniquely identify storage objects.
func makeKey(rlsname string, version int) string {
	return fmt.Sprintf("%s.v%d", rlsname, version)
}
