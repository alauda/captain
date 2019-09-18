/*
Copyright 2017 The Kubernetes Authors.

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

package controller

import (
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/alauda/captain/pkg/cluster"
	"github.com/alauda/captain/pkg/config"
	"github.com/alauda/captain/pkg/helm"
	"github.com/alauda/captain/pkg/util"
	funk "github.com/thoas/go-funk"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"

	"github.com/alauda/component-base/hash"

	commoncache "github.com/patrickmn/go-cache"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	alpha1 "github.com/alauda/helm-crds/pkg/apis/app/v1alpha1"
	clientset "github.com/alauda/helm-crds/pkg/client/clientset/versioned"

	informers "github.com/alauda/helm-crds/pkg/client/informers/externalversions"
	listers "github.com/alauda/helm-crds/pkg/client/listers/app/v1alpha1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	clusterclientset "k8s.io/cluster-registry/pkg/client/clientset/versioned"
)

type clusterConfig struct {
	// clusterClient is used to access Cluster Resource
	clusterClient clusterclientset.Interface

	// clusterNamespace is the namespace that all the Cluster resource lives in
	clusterNamespace string

	globalClusterName string
}

// Controller is the controller implementation for HelmRequest resources
type Controller struct {
	// kubeClient is a standard kubernetes clientset
	// usage:
	// 1. retrieve cluster admin token from secret
	kubeClient kubernetes.Interface
	// appClientSet is a clientset for our own API group
	appClientSet clientset.Interface

	clusterConfig clusterConfig

	// restConfig is the kubernetes rest config for the current cluster, used for
	// sync HelmRequest who's cluster name is "".
	restConfig *rest.Config

	helmRequestLister listers.HelmRequestLister
	helmRequestSynced cache.InformerSynced

	chartRepoSynced cache.InformerSynced

	// ClusterCache is used to store Cluster resource
	ClusterCache *commoncache.Cache

	// workQueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workQueue workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder
}

//NewController create a new controller
func NewController(mgr manager.Manager, opt *config.Options, stopCh <-chan struct{}) (*Controller, error) {
	cfg := mgr.GetConfig()
	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	appClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	clusterClient, err := clusterclientset.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	// kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)
	appInformerFactory := informers.NewSharedInformerFactory(appClient, time.Second*30)
	chartRepoInformerFactory := informers.NewSharedInformerFactoryWithOptions(appClient, time.Second*30, informers.WithNamespace(opt.ChartRepoNamespace))

	informer := appInformerFactory.App().V1alpha1().HelmRequests()
	repoInformer := chartRepoInformerFactory.App().V1alpha1().ChartRepos()

	controller := &Controller{
		kubeClient:   kubeClient,
		appClientSet: appClient,
		clusterConfig: clusterConfig{
			clusterNamespace:  opt.ClusterNamespace,
			clusterClient:     clusterClient,
			globalClusterName: opt.GlobalClusterName,
		},
		restConfig:        cfg,
		recorder:          mgr.GetEventRecorderFor(util.ComponentName),
		helmRequestLister: informer.Lister(),
		helmRequestSynced: informer.Informer().HasSynced,
		chartRepoSynced:   repoInformer.Informer().HasSynced,
		workQueue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "HelmRequests"),
		// refresh frequently
		ClusterCache: commoncache.New(1*time.Minute, 5*time.Minute),
	}

	klog.Info("Setting up event handlers")
	// Set up an event handler for when HelmRequest resources change
	informer.Informer().AddEventHandler(controller.newHelmRequestHandler())
	repoInformer.Informer().AddEventHandler(controller.newChartRepoHandler())

	klog.V(7).Infof("cluster rest config is : %+v", cfg)

	// notice that there is no need to run Start methods in a separate goroutine. (i.e. go kubeInformerFactory.Start(stopCh)
	// Start method is non-blocking and runs all registered informers in a dedicated goroutine.
	// kubeInformerFactory.Start(stopCh)
	// appInformerFactory.Start(stopCh)

	// fuck examples, this should after init controller
	appInformerFactory.Start(stopCh)
	chartRepoInformerFactory.Start(stopCh)

	return controller, mgr.Add(controller)
}

//GetClusterClient get a client for access Cluster resource
func (c *Controller) GetClusterClient() clusterclientset.Interface {
	return c.clusterConfig.clusterClient
}

// Start will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workQueue and wait for
// workers to finish processing their current work items.
func (c *Controller) Start(stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workQueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	klog.Info("Starting HelmRequest controller")

	// Wait for the caches to be synced before starting workers
	klog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.helmRequestSynced, c.chartRepoSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	klog.Info("Starting workers")
	// Launch two workers to process HelmRequest resources
	for i := 0; i < 2; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	klog.Info("Started workers")
	<-stopCh
	klog.Info("Shutting down workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workQueue.
func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workQueue and
// attempt to process it, by calling the syncHandler.
func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workQueue.Get()

	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workQueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workQueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workQueue and attempted again after a back-off
		// period.
		defer c.workQueue.Done(obj)
		var key string
		var ok bool
		// We expect strings to come off the workQueue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// workQueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the
		// workQueue.
		if key, ok = obj.(string); !ok {
			// As the item in the workQueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workQueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workQueue but got %#v", obj))
			return nil
		}
		// Start the syncHandler, passing it the namespace/name string of the
		// HelmRequest resource to be synced.
		if err := c.syncHandler(key); err != nil {
			// Put the item back on the workQueue to handle any transient errors.
			c.workQueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workQueue.Forget(obj)
		klog.Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the HelmRequest resource
// with the current status of the resource.
func (c *Controller) syncHandler(key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}
	klog.V(9).Infof("")

	// Get the HelmRequest resource with this namespace/name
	helmRequest, err := c.helmRequestLister.HelmRequests(namespace).Get(name)
	if err != nil {
		// The HelmRequest resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("helmRequest '%s' in work queue no longer exists", key))
			return nil
		}

		return err
	}

	if !helmRequest.DeletionTimestamp.IsZero() {
		klog.Infof("HelmRequest has not nil DeletionTimestamp, starting to delete it: %s", helmRequest.Name)
		if err := c.deleteHelmRequest(helmRequest); err != nil {
			c.sendFailedDeleteEvent(helmRequest, err)
			return err
		}
		return nil
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
			if helmRequest.Status.Phase != alpha1.HelmRequestSynced {
				klog.Infof("helm request phase not synced, trying to set it")
				return c.updateHelmRequestPhase(helmRequest, alpha1.HelmRequestSynced)
			}
			return nil
		}
		klog.Infof("sync HelmRequest %s to cluster %s", key, helmRequest.Spec.ClusterName)
		if err := c.syncToCluster(helmRequest); err != nil {
			c.setSyncFailedStatus(helmRequest, err)
			return err
		}
	} else if err := c.syncToAllClusters(key, helmRequest); err != nil {
		c.setSyncFailedStatus(helmRequest, err)
		return err
	}

	c.recorder.Event(helmRequest, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil
}

// see issue: https://github.com/kubernetes/kubernetes/issues/60845
// the origin code in sample-controller is not working... fuck
func (c *Controller) updateHelmRequestStatus(helmRequest *alpha1.HelmRequest) error {
	// NEVER modify objects from the store. It's a read-only, local cache.
	// You can use DeepCopy() to make a deep copy of original object and modify this copy
	// Or create a copy manually for better performance
	h := hash.GenHashStr(helmRequest.Spec)
	// Note: we have to generate the hash before the deepcopy, because somehow the deepcopy
	// can create a spec that have different hash value.
	request := helmRequest.DeepCopy()
	request.Status.LastSpecHash = h
	return c.updateHelmRequestPhase(request, alpha1.HelmRequestSynced)
}

// setPartialSyncedStatus set spec hash and partial-synced status for helm-request
//TODO: merge with updateHelmRequestStatus
func (c *Controller) setPartialSyncedStatus(helmRequest *alpha1.HelmRequest) error {
	h := hash.GenHashStr(helmRequest.Spec)
	request := helmRequest.DeepCopy()
	request.Status.LastSpecHash = h
	return c.updateHelmRequestPhase(request, alpha1.HelmRequestPartialSynced)
}

// setSyncFailedStatus set HelmRequest to Failed and generated a warning event
func (c *Controller) setSyncFailedStatus(helmRequest *alpha1.HelmRequest, err error) error {
	c.sendFailedSyncEvent(helmRequest, err)
	return c.updateHelmRequestPhase(helmRequest, alpha1.HelmRequestFailed)

}

func (c *Controller) updateHelmRequestPhase(helmRequest *alpha1.HelmRequest, phase alpha1.HelmRequestPhase) error {
	request := helmRequest.DeepCopy()
	request.Status.Phase = phase

	// If the CustomResourceSubresources feature gate is not enabled,
	// we must use Update instead of UpdateStatus to update the Status block of the HelmRequest resource.
	// UpdateStatus will not allow changes to the Spec of the resource,
	// which is ideal for ensuring nothing other than resource status has been updated.
	_, err := c.appClientSet.AppV1alpha1().HelmRequests(helmRequest.Namespace).UpdateStatus(request)
	if err != nil {
		if apierrors.IsConflict(err) {
			klog.Warning("update helm request status conflict, retry...")
			origin, err := c.appClientSet.AppV1alpha1().HelmRequests(helmRequest.Namespace).Get(helmRequest.Name, metav1.GetOptions{})
			if err != nil {
				return err
			}
			klog.Warningf("origin status: %+v, current: %+v", origin.Status, request.Status)
			origin.Status = *request.Status.DeepCopy()
			_, err = c.appClientSet.AppV1alpha1().HelmRequests(helmRequest.Namespace).UpdateStatus(origin)
			if err != nil {
				klog.Error("retrying update helmrequest status error:", err)
			}
			return err
		}
		klog.Errorf("update status for helmrequest %s error: %s", helmRequest.Name, err.Error())
	}
	return err
}

// removeFinalizer remove all the finalizers of this HelmRequest
func (c *Controller) removeFinalizer(helmRequest *alpha1.HelmRequest) error {
	if funk.Contains(helmRequest.Finalizers, util.FinalizerName) {
		klog.Infof("found finalizers for helmrequest: %s", helmRequest.Name)
		data := `{"metadata":{"finalizers":null}}`
		// ? only patch can work?
		_, err := c.appClientSet.AppV1alpha1().HelmRequests(helmRequest.Namespace).Patch(
			helmRequest.Name, types.MergePatchType, []byte(data),
		)
		return err
	}
	return nil
}

// enqueueHelmRequest takes a HelmRequest resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than HelmRequest.
func (c *Controller) enqueueHelmRequest(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workQueue.Add(key)
}

func (c *Controller) deleteHandler(obj interface{}) {
	var err error
	var key string
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}

	hr := obj.(*alpha1.HelmRequest)

	err = c.deleteHelmRequest(hr)
	if err != nil {
		c.sendFailedDeleteEvent(hr, err)
		utilruntime.HandleError(err)
		c.workQueue.AddRateLimited(key)
	} else {
		c.recorder.Event(hr, corev1.EventTypeNormal, SuccessfulDelete,
			fmt.Sprintf("Deleted HelmRequest: %s", hr.GetName()))
	}

}

// deleteHelmRequest delete the installed chart about this HelmRequest
// if InstallToAllClusters=true, delete it from all clusters
func (c *Controller) deleteHelmRequest(hr *alpha1.HelmRequest) error {
	// get clusters
	var clusters []*cluster.Info
	if hr.Spec.InstallToAllClusters {
		result, err := c.getAllClusters()
		if err != nil {
			return err
		}
		clusters = result
	} else {
		info, err := c.getClusterInfo(hr.Spec.ClusterName)
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
		ci.Namespace = hr.Spec.Namespace
		klog.Infof("delete HelmRequest %s for cluster %s", hr.GetName(), ci.Name)
		err := helm.Delete(hr, &ci)
		if err != nil {
			errs = append(errs, err)
		}
	}

	err := utilerrors.NewAggregate(errs)
	if err != nil {
		return err
	}

	if err := c.removeFinalizer(hr); err != nil {
		return err
	}
	klog.Infof("successfully remove finalizers from helmrequest: %s", hr.Name)

	return nil
}
