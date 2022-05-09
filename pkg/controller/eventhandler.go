package controller

import (
	"encoding/json"
	"reflect"

	appv1 "github.com/alauda/helm-crds/pkg/apis/app/v1"
	appv1alpha1 "github.com/alauda/helm-crds/pkg/apis/app/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
)

// newHelmRequestHandler create an HelmRequest handler
// The cache package have a built-in filter-support handler, but it cannot compare the old and new
// obj at the same time
func (c *Controller) newHelmRequestHandler() cache.ResourceEventHandler {
	updateFunc := func(old, new interface{}) {
		oldHR, ok := old.(*appv1.HelmRequest)
		if !ok {
			var err error
			oldHR, err = convertToV1(old)
			if err != nil {
				klog.Errorf("can not convert object to v1 helmrequest : %+v", old)
				return
			}
		}
		newHR, ok := new.(*appv1.HelmRequest)
		if !ok {
			var err error
			newHR, err = convertToV1(new)
			if err != nil {
				klog.Errorf("can not convert object to v1 helmrequest : %+v", new)
				return
			}
		}
		// this is a bit of tricky
		// 1. old and new -> 1 cluster => check version and spec
		// 2. old and new -> N cluster => no check
		// 3. old 1 / new N => spec and version changed
		// 4. old N / new 1 => spec and version changed

		if oldHR.Spec.InstallToAllClusters && newHR.Spec.InstallToAllClusters {
			c.enqueueHelmRequest(new)
		} else {

			if newHR.DeletionTimestamp != nil {
				klog.V(4).Infof("get an helmrequest with deletiontimestap: %s", newHR.Name)
				c.enqueueHelmRequest(new)
			}

			if oldHR.ResourceVersion == newHR.ResourceVersion {
				return
			}

			if reflect.DeepEqual(oldHR.Spec, newHR.Spec) && reflect.DeepEqual(oldHR.Annotations, newHR.Annotations) {
				klog.V(4).Infof("spec/annotations equal, not update: %s", newHR.Name)
				return
			}
			klog.V(4).Infof("old hr annotations: %+v, new hr annotations: %+v", oldHR.Annotations, newHR.Annotations)
			klog.V(4).Infof("old hr.spec: %+v, new hr.spec: %+v", oldHR.Spec, newHR.Spec)
			c.enqueueHelmRequest(new)
		}
	}

	funcs := cache.ResourceEventHandlerFuncs{
		AddFunc:    c.enqueueHelmRequest,
		UpdateFunc: updateFunc,
		DeleteFunc: c.deleteHandler,
	}

	return funcs
}

// newClusterHelmRequestHandler create an HelmRequest handler
// The cache package have a built-in filter-support handler, but it cannot compare the old and new
// obj at the same time
func (c *Controller) newClusterHelmRequestHandler(name string) cache.ResourceEventHandler {
	updateFunc := func(old, new interface{}) {
		oldHR, ok := old.(*appv1.HelmRequest)
		if !ok {
			var err error
			oldHR, err = convertToV1(old)
			if err != nil {
				klog.Errorf("can not convert object to v1 helmrequest : %+v", old)
				return
			}
		}
		newHR, ok := new.(*appv1.HelmRequest)
		if !ok {
			var err error
			newHR, err = convertToV1(new)
			if err != nil {
				klog.Errorf("can not convert object to v1 helmrequest : %+v", new)
				return
			}
		}
		// this is a bit of tricky
		// 1. old and new -> 1 cluster => check version and spec
		// 2. old and new -> N cluster => no check
		// 3. old 1 / new N => spec and version changed
		// 4. old N / new 1 => spec and version changed

		if oldHR.Spec.InstallToAllClusters && newHR.Spec.InstallToAllClusters {
			c.enqueueClusterHelmRequest(new, name)
		} else {

			if newHR.DeletionTimestamp != nil {
				klog.V(4).Infof("get an helmrequest with deletiontimestap: %s", newHR.Name)
				c.enqueueClusterHelmRequest(new, name)
			}

			if newHR.Status.Phase == appv1.HelmRequestPending {
				klog.V(4).Infof("")
			}

			if oldHR.ResourceVersion == newHR.ResourceVersion {
				return
			}
			if reflect.DeepEqual(oldHR.Spec, newHR.Spec) {
				klog.V(4).Infof("spec equal, not update: %s", newHR.Name)
				return
			}
			klog.V(4).Infof("old hr: %+v, new hr: %+v", oldHR.Spec, newHR.Spec)
			c.enqueueClusterHelmRequest(new, name)
		}
	}

	addFunc := func(obj interface{}) {
		klog.Infof("receive hr create event: %+v", obj)
		c.enqueueClusterHelmRequest(obj, name)
	}

	deleteFunc := func(obj interface{}) {
		c.deleteClusterHelmRequestHandler(obj, name)
	}

	funcs := cache.ResourceEventHandlerFuncs{
		AddFunc:    addFunc,
		UpdateFunc: updateFunc,
		DeleteFunc: deleteFunc,
	}

	return funcs
}

func convertToV1(obj interface{}) (*appv1.HelmRequest, error) {
	alphaHR := obj.(*appv1alpha1.HelmRequest)
	gvk := schema.GroupVersionKind{
		Group:   appv1.SchemeGroupVersion.Group,
		Version: appv1.SchemeGroupVersion.Version,
		Kind:    alphaHR.GetObjectKind().GroupVersionKind().Kind,
	}
	alphaHR.SetGroupVersionKind(gvk)

	b, err := json.Marshal(alphaHR)
	if err != nil {
		return nil, err
	}

	convertHR := &appv1.HelmRequest{}
	if err := json.Unmarshal(b, convertHR); err != nil {
		return nil, err
	}

	return convertHR, nil
}
