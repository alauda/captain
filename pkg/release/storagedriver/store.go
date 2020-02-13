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
	"strconv"
	"strings"
	"time"

	"github.com/alauda/helm-crds/pkg/apis/app/v1alpha1"
	releaseclient "github.com/alauda/helm-crds/pkg/client/clientset/versioned/typed/app/v1alpha1"
	"github.com/pkg/errors"
	rspb "helm.sh/helm/pkg/release"
	"helm.sh/helm/pkg/storage/driver"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kblabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/klog"
)

var _ driver.Driver = (*Releases)(nil)

// ReleasesDriverName is the string name of the driver.
const ReleasesDriverName = "Release"

// Releases is a wrapper around an implementation of a kubernetes
// ReleasesInterface.
type Releases struct {
	impl releaseclient.ReleaseInterface
	Log  func(string, ...interface{})
}

// NewReleases initializes a new Releases wrapping an implementation of
// the kubernetes ReleasesInterface.
func NewReleases(impl releaseclient.ReleaseInterface) *Releases {
	return &Releases{
		impl: impl,
		Log:  klog.Infof,
	}
}

// Name returns the name of the driver.
func (rel *Releases) Name() string {
	return ReleasesDriverName
}

func (rel *Releases) getRawRelease(key string) (*v1alpha1.Release, error) {
	obj, err := rel.impl.Get(key, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, driver.ErrReleaseNotFound
		}

		rel.Log("get: failed to get %q: %s", key, err)
		return nil, err
	}
	return obj, err
}

// Get fetches the release named by key. The corresponding release is returned
// or error if not found.
func (rel *Releases) Get(key string) (*rspb.Release, error) {
	// fetch the configmap holding the release named by key
	obj, err := rel.getRawRelease(key)
	if err != nil {
		return nil, err
	}
	// found the configmap, decode the base64 data string
	r, err := decodeRelease(obj)
	if err != nil {
		rel.Log("get: failed to decode data %q: %s", key, err)
		return nil, err
	}
	// return the release object
	return r, nil
}

// List fetches all releases and returns the list releases such
// that filter(release) == true. An error is returned if the
// configmap fails to retrieve the releases.
func (rel *Releases) List(filter func(*rspb.Release) bool) ([]*rspb.Release, error) {
	lsel := kblabels.Set{"owner": "helm"}.AsSelector()
	opts := metav1.ListOptions{LabelSelector: lsel.String()}

	list, err := rel.impl.List(opts)
	if err != nil {
		rel.Log("list: failed to list: %s", err)
		return nil, err
	}

	var results []*rspb.Release

	// iterate over the configmaps object list
	// and decode each release
	for _, item := range list.Items {
		rls, err := decodeRelease(&item)
		if err != nil {
			rel.Log("list: failed to decode release: %v: %s", item, err)
			continue
		}
		if filter(rls) {
			results = append(results, rls)
		}
	}
	return results, nil
}

// Query fetches all releases that match the provided map of labels.
// An error is returned if the configmap fails to retrieve the releases.
func (rel *Releases) Query(labels map[string]string) ([]*rspb.Release, error) {
	ls := kblabels.Set{}
	for k, v := range labels {
		if errs := validation.IsValidLabelValue(v); len(errs) != 0 {
			return nil, errors.Errorf("invalid label value: %q: %s", v, strings.Join(errs, "; "))
		}
		ls[k] = v
	}

	opts := metav1.ListOptions{LabelSelector: ls.AsSelector().String()}

	list, err := rel.impl.List(opts)
	if err != nil {
		rel.Log("query: failed to query with labels: %s", err)
		return nil, err
	}

	if len(list.Items) == 0 {
		return nil, driver.ErrReleaseNotFound
	}

	var results []*rspb.Release
	for _, item := range list.Items {
		rls, err := decodeRelease(&item)
		if err != nil {
			rel.Log("query: failed to decode release: %s", err)
			continue
		}
		results = append(results, rls)
	}
	return results, nil
}

// Create creates a new ConfigMap holding the release. If the
// ConfigMap already exists, ErrReleaseExists is returned.
func (rel *Releases) Create(key string, rls *rspb.Release) error {
	// set labels for configmaps object meta data
	var lbs labels

	lbs.init()
	lbs.set("createdAt", strconv.Itoa(int(time.Now().Unix())))

	// create a new configmap to hold the release
	obj, err := newReleasesObject(key, rls, lbs)
	if err != nil {
		rel.Log("create: failed to encode release %q: %s", rls.Name, err)
		return err
	}
	// push the configmap object out into the kubiverse
	if _, err := rel.impl.Create(obj); err != nil {
		if apierrors.IsAlreadyExists(err) {
			return driver.ErrReleaseExists
		}

		rel.Log("create: failed to create: %s", err)
		return err
	}
	return nil
}

// Update updates the ConfigMap holding the release. If not found
// the ConfigMap is created to hold the release.
func (rel *Releases) Update(key string, rls *rspb.Release) error {
	// set labels for configmaps object meta data
	var lbs labels

	lbs.init()
	lbs.set("modifiedAt", strconv.Itoa(int(time.Now().Unix())))

	// create a new configmap object to hold the release
	obj, err := newReleasesObject(key, rls, lbs)
	if err != nil {
		rel.Log("update: failed to encode release %q: %s", rls.Name, err)
		return err
	}

	old, err := rel.getRawRelease(key)
	if err != nil {
		rel.Log("update, pre-fetch release error: %s: %s", key, err.Error())
	} else {
		obj.ResourceVersion = old.ResourceVersion
	}

	// push the configmap object out into the kubiverse
	_, err = rel.impl.Update(obj)
	if err != nil {
		rel.Log("update: failed to update: %s", err)
		return err
	}
	return nil
}

// Delete deletes the ConfigMap holding the release named by key.
func (rel *Releases) Delete(key string) (rls *rspb.Release, err error) {
	// fetch the release to check existence
	if rls, err = rel.Get(key); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, driver.ErrReleaseExists
		}

		rel.Log("delete: failed to get release %q: %s", key, err)
		return nil, err
	}
	// delete the release
	if err = rel.impl.Delete(key, &metav1.DeleteOptions{}); err != nil {
		return rls, err
	}
	return rls, nil
}

// newReleasesObject constructs a kubernetes ConfigMap object
// to store a release. Each configmap data entry is the base64
// encoded string of a release's binary protobuf encoding.
//
// The following labels are used within each configmap:
//
//    "modifiedAt"     - timestamp indicating when this configmap was last modified. (set in Update)
//    "createdAt"      - timestamp indicating when this configmap was created. (set in Create)
//    "version"        - version of the release.
//    "status"         - status of the release (see proto/hapi/release.status.pb.go for variants)
//    "owner"          - owner of the configmap, currently "helm".
//    "name"           - name of the release.
//
func newReleasesObject(key string, rls *rspb.Release, lbs labels) (*v1alpha1.Release, error) {
	const owner = "helm"

	// encode the release
	s, err := encodeRelease(rls)
	if err != nil {
		return nil, err
	}

	if lbs == nil {
		lbs.init()
	}

	// apply labels
	lbs.set("name", rls.Name)
	lbs.set("owner", owner)
	lbs.set("status", rls.Info.Status.String())
	lbs.set("version", strconv.Itoa(rls.Version))

	s.Labels = lbs.toMap()
	s.Name = key

	return s, nil
}
