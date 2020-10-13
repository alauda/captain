package controller

import (
	"fmt"
	"strings"

	"github.com/alauda/captain/pkg/cluster"
	"github.com/alauda/captain/pkg/helm"
	"github.com/alauda/captain/pkg/util"
	"github.com/alauda/helm-crds/pkg/apis/app/v1alpha1"
	"github.com/thoas/go-funk"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
)

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the HelmRequest resource
// with the current status of the resource.
// key has two format:
// 1. <namespace>/<name>
// 2. <cluster>/<namespace>/<name>
func (c *Controller) syncHandler(key string) error {
	klog.Infof("Start sync helmrequest: %s", key)
	clusterName, key := splitClusterKey(key)

	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the HelmRequest resource with this namespace/name
	helmRequest, err := c.getHelmRequestLister(clusterName).HelmRequests(namespace).Get(name)
	if err != nil {
		// The HelmRequest resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			runtime.HandleError(fmt.Errorf("helmRequest '%s' in work queue no longer exists", key))
			return nil
		}

		return err
	}

	if helmRequest.Annotations != nil {
		if helmRequest.Annotations[util.NoSyncAnotation] == "true" {
			klog.Info("found helmrequest want to be ignored by captain: ", helmRequest.Name)
			return nil
		}
	}

	helmRequest.ClusterName = clusterName

	if !helmRequest.DeletionTimestamp.IsZero() {
		klog.Infof("HelmRequest has not nil DeletionTimestamp, starting to delete it: %s", helmRequest.Name)
		if err := c.deleteHelmRequest(helmRequest); err != nil {
			c.sendFailedDeleteEvent(helmRequest, err)
			return err
		}
		return nil
	}

	// add finalizer if needed, cluster name is already set
	if err := c.addFinalizer(helmRequest); err != nil {
		klog.Errorf("add finalizer for helmrequest %s error, err is: %+v", helmRequest.Name, err)
		c.sendFailedSyncEvent(helmRequest, err)
		return err
	}

	// check dependencies
	if err := c.checkDependenciesForHelmRequest(helmRequest); err != nil {
		klog.Infof("check dependencies for %s not pass, err is : %+v", helmRequest.Name, err)
		c.sendFailedSyncEvent(helmRequest, err)
		return err
	}

	klog.Infof("dependency check pass for HelmRequest %s", helmRequest.GetName())

	if !helmRequest.Spec.InstallToAllClusters {

		if helm.IsHelmRequestSynced(helmRequest) {
			klog.Infof("HelmRequest %s synced", helmRequest.Name)
			if helmRequest.Status.Phase != v1alpha1.HelmRequestSynced {
				klog.Infof("helm request phase not synced, trying to set it")
				return c.updateHelmRequestPhase(helmRequest, v1alpha1.HelmRequestSynced)
			}
			return nil
		}
		c.setPendingStatus(helmRequest)
		klog.Infof("sync HelmRequest %s to cluster %s", key, helmRequest.Spec.ClusterName)
		if err := c.syncToCluster(helmRequest); err != nil {
			c.setSyncFailedStatus(helmRequest, err)
			return err
		}
	} else if err := c.syncToAllClusters(key, helmRequest); err != nil {
		c.setSyncFailedStatus(helmRequest, err)
		return err
	}

	// If we send event here, HelmRequest enabled installToAllCluster will send
	c.getEventRecorder(helmRequest).Event(helmRequest, v1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil
}

// syncToCluster install/update HelmRequest to one cluster
func (c *Controller) syncToCluster(helmRequest *v1alpha1.HelmRequest) error {
	clusterName := c.getDeployCluster(helmRequest)
	info, err := c.getClusterInfo(clusterName)
	if err != nil {
		klog.Errorf("get cluster info error: %s", err.Error())
		return err
	}

	klog.Infof("get cluster %s  endpoint: %s", info.Name, info.Endpoint)

	if err := c.sync(info, helmRequest); err != nil {
		return err
	}

	// Finally, we update the status block of the HelmRequest resource to reflect the
	// current state of the world
	err = c.updateHelmRequestStatus(helmRequest)
	return err
}

// enqueueHelmRequest takes a HelmRequest resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than HelmRequest.
func (c *Controller) enqueueHelmRequest(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	c.workQueue.Add(key)
}

// isOldEvent filters outdated event by compare resource uid
// in case the resource not exist, we continue to real delete.
func (c *Controller) isOldEvent(cluster string, hr *v1alpha1.HelmRequest) (bool, error) {
	// Get the HelmRequest resource with this namespace/name
	// hr should have clusterName set.
	current, err := c.getAppClient(hr).AppV1alpha1().HelmRequests(hr.Namespace).Get(hr.Name, metav1.GetOptions{})
	// don't want use the cached one.
	// current, err := c.getHelmRequestLister(cluster).HelmRequests(hr.Namespace).Get(hr.Name)
	if err != nil {
		// The HelmRequest resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			klog.Info("helmrequest not found when check version for delete, ignore")
			return false, nil
		}
		return false, err
	}

	if current.UID != hr.UID {
		klog.Warningf("received old delete event for helmrequest: %s %s %s", hr.Name, hr.UID, current.UID)
		return true, nil
	}

	//received := cast.ToInt(hr.ResourceVersion)
	//exist := cast.ToInt(current.ResourceVersion)
	//
	//if received < exist {
	//	klog.Warningf("received old delete event for helmrequest: %s %d %d", hr.Name, received, exist)
	//	return true, nil
	//}
	//
	//receivedTimestamp := hr.GetCreationTimestamp()
	//existTimestamp := current.GetCreationTimestamp()
	//if receivedTimestamp.Before(&existTimestamp) {
	//	klog.Warningf("received old delete event for helmrequest: %s %s %s", hr.Name, receivedTimestamp.String(), existTimestamp.String())
	//}

	return false, nil
}

// deleteHandler is delete handler for HelmRequest in global cluster
func (c *Controller) deleteHandler(obj interface{}) {
	var err error
	var key string
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}

	hr := obj.(*v1alpha1.HelmRequest)

	klog.Infof("receive delete event: %+v", hr)

	outdated, err := c.isOldEvent("", hr)
	if err != nil {
		c.sendFailedDeleteEvent(hr, err)
		runtime.HandleError(err)
		c.workQueue.AddRateLimited(key)
		return
	}

	if outdated {
		return
	}

	err = c.deleteHelmRequest(hr)
	if err != nil {
		c.sendFailedDeleteEvent(hr, err)
		runtime.HandleError(err)
		c.workQueue.AddRateLimited(key)
	} else {
		c.getEventRecorder(hr).Event(hr, v1.EventTypeNormal, SuccessfulDelete,
			fmt.Sprintf("Deleted HelmRequest: %s", hr.GetName()))
	}
}

// deleteHelmRequest delete the installed chart about this HelmRequest
// if InstallToAllClusters=true, delete it from all clusters
func (c *Controller) deleteHelmRequest(hr *v1alpha1.HelmRequest) error {
	// get clusters
	var clusters []*cluster.Info
	if hr.Spec.InstallToAllClusters {
		result, err := c.getAllClusters()
		if err != nil {
			return err
		}
		clusters = result
	} else {
		// no longer use .spec.ClusterName
		info, err := c.getClusterInfo(c.getDeployCluster(hr))
		if err != nil {
			if errors.IsNotFound(err) {
				klog.Warning("cluster not found when delete helmrequest, ignore it")
				return c.removeFinalizer(hr)
			}
			return err
		}
		clusters = append(clusters, info)
	}

	var errs []error

	// loop to delete in all clusters
	for _, info := range clusters {
		ci := *info
		ci.Namespace = hr.GetReleaseNamespace()
		klog.Infof("delete HelmRequest %s for cluster %s", hr.GetName(), ci.Name)

		d := helm.NewDeploy()
		d.HelmRequest = hr
		d.Cluster = &ci

		err := d.Delete()
		if err != nil {
			errs = append(errs, err)
		}
	}

	err := utilerrors.NewAggregate(errs)
	if err != nil {
		// if this error was caused by `build resource error`, which is usually caused by CRD issue,
		// this will block the delete forever, so we just ignore this error, since most of the time,
		// there are no resource created
		if strings.Contains(err.Error(), "unable to build kubernetes objects for delete") {
			klog.Warning("unable to build kubernetes resource when delete, ignore this error: ", err)
		} else {
			return err
		}

	}

	if err := c.removeFinalizer(hr); err != nil {
		return err
	}
	klog.Infof("successfully remove finalizers from helmrequest: %s", hr.Name)

	return nil
}

// addFinalizer add finalizer to a hr
// 1. support add finalizer for global/business clusters
// 2. if there is and old hr without finalizer, compare uid of the event and only delete it if only the uid match
// 3. hr should carry with cluster info
func (c *Controller) addFinalizer(hr *v1alpha1.HelmRequest) error {
	if !funk.Contains(hr.Finalizers, util.FinalizerName) {
		data := `{"metadata":{"finalizers":["captain.cpaas.io"]}}`
		_, err := c.getAppClient(hr).AppV1alpha1().HelmRequests(hr.Namespace).Patch(
			hr.Name, types.MergePatchType, []byte(data),
		)
		return err
	}
	return nil

}

// removeFinalizer remove all the finalizers of this HelmRequest
func (c *Controller) removeFinalizer(helmRequest *v1alpha1.HelmRequest) error {
	// captain.alauda.io is the old finalizer, this should provide some backword compatibility
	if funk.Contains(helmRequest.Finalizers, util.FinalizerName) || funk.Contains(helmRequest.Finalizers, "captain.alauda.io") {
		klog.Infof("found finalizers for helmrequest: %s", helmRequest.Name)
		data := `{"metadata":{"finalizers":null}}`
		// ? only patch can work?
		_, err := c.getAppClient(helmRequest).AppV1alpha1().HelmRequests(helmRequest.Namespace).Patch(
			helmRequest.Name, types.MergePatchType, []byte(data),
		)
		return err
	}
	return nil
}

// updateHelmRequestPhase update a helmrequest status
// If this helmrequest not exist already(delete by user, remove the release)
func (c *Controller) updateHelmRequestPhase(helmRequest *v1alpha1.HelmRequest, phase v1alpha1.HelmRequestPhase) error {
	request := helmRequest.DeepCopy()
	request.Status.Phase = phase

	client := c.getAppClient(helmRequest)

	// If the CustomResourceSubresources feature gate is not enabled,
	// we must use Update instead of UpdateStatus to update the Status block of the HelmRequest resource.
	// UpdateStatus will not allow changes to the Spec of the resource,
	// which is ideal for ensuring nothing other than resource status has been updated.
	_, err := client.AppV1alpha1().HelmRequests(helmRequest.Namespace).UpdateStatus(request)
	if err != nil {
		if apierrors.IsConflict(err) {
			klog.Warning("update helm request status conflict, retry...")
			origin, err := client.AppV1alpha1().HelmRequests(helmRequest.Namespace).Get(helmRequest.Name, metav1.GetOptions{})
			if err != nil {
				if apierrors.IsNotFound(err) {
					klog.Warning("helmrequest not found when trying to update status, delete the release...", helmRequest.Name)
					return c.deleteHelmRequest(helmRequest)
				}
				return err
			}
			klog.Warningf("origin status: %+v, current: %+v", origin.Status, request.Status)
			origin.Status = *request.Status.DeepCopy()
			_, err = client.AppV1alpha1().HelmRequests(helmRequest.Namespace).UpdateStatus(origin)
			if err != nil {
				klog.Error("retrying update helmrequest status error:", err)
			}
			return err
		}
		klog.Errorf("update status for helmrequest %s error: %s", helmRequest.Name, err.Error())
	}
	return err
}
