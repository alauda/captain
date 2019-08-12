package controller

import (
	"reflect"

	"github.com/alauda/captain/pkg/helm"

	alpha1 "github.com/alauda/captain/pkg/apis/app/v1alpha1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
)

// newChartRepoHandler create new event handler for ChartRepo
// Because we only need to watch ChartRepos from one namespace, for now we will
// use a dumb filter...
func (c *Controller) newChartRepoHandler() cache.ResourceEventHandler {
	updateFunc := func(old, new interface{}) {
		// oldChartRepo := old.(*alpha1.ChartRepo)
		// newChartRepo := new.(*alpha1.ChartRepo)

		c.syncChartRepo(new)

		//if oldChartRepo.Spec.Secret != nil && newChartRepo.Spec.Secret != nil {
		//	if oldChartRepo.Spec.Secret.Name != newChartRepo.Spec.Secret.Name {
		//		c.syncChartRepo(newChartRepo)
		//		return
		//	}
		//	if oldChartRepo.Spec.Secret.Namespace != newChartRepo.Spec.Secret.Namespace {
		//		c.syncChartRepo(newChartRepo)
		//		return
		//	}
		//}
	}

	deleteFunc := func(obj interface{}) {
		cr, ok := obj.(*alpha1.ChartRepo)
		if ok {
			if err := helm.RemoveRepository(cr.GetName()); err != nil {
				klog.Error("remove repo from helm error:", err)
			}
		}
	}

	funcs := cache.ResourceEventHandlerFuncs{
		AddFunc:    c.syncChartRepo,
		UpdateFunc: updateFunc,
		DeleteFunc: deleteFunc,
	}

	return funcs
}

// newHelmRequestHandler create an HelmRequest handler
// The cache package have a built-in filter-support handler, but it cannot compare the old and new
// obj at the same time
func (c *Controller) newHelmRequestHandler() cache.ResourceEventHandler {
	updateFunc := func(old, new interface{}) {
		oldHR := old.(*alpha1.HelmRequest)
		newHR := new.(*alpha1.HelmRequest)
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
			if reflect.DeepEqual(oldHR.Spec, newHR.Spec) {
				klog.V(4).Infof("spec equal, not update: %s", newHR.Name)
				return
			}
			klog.V(4).Infof("old hr: %+v, new hr: %+v", oldHR.Spec, newHR.Spec)
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
