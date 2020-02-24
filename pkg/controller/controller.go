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
	"github.com/alauda/captain/pkg/helm"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/alauda/captain/pkg/config"
	"github.com/alauda/captain/pkg/util"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"

	alpha1 "github.com/alauda/helm-crds/pkg/apis/app/v1alpha1"
	clientset "github.com/alauda/helm-crds/pkg/client/clientset/versioned"
	commoncache "github.com/patrickmn/go-cache"

	informers "github.com/alauda/helm-crds/pkg/client/informers/externalversions"
	listers "github.com/alauda/helm-crds/pkg/client/listers/app/v1alpha1"
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

	// this is where all the ChartRepo/Charts lives
	systemNamespace string

	// restConfig is the kubernetes rest config for the current cluster, used for
	// sync HelmRequest who's cluster name is "".
	restConfig *rest.Config

	helmRequestLister listers.HelmRequestLister
	helmRequestSynced cache.InformerSynced

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

	// To support multiple cluster, we have to watch all the clusters for HelmRequests
	// May be we should remove the old field for global cluster...
	clusterHelmRequestListers map[string]listers.HelmRequestLister
	clusterHelmRequestSynced  map[string]cache.InformerSynced
	clusterWorkQueues         map[string]workqueue.RateLimitingInterface
	clusterClients            map[string]clientset.Interface
	clusterRecorders          map[string]record.EventRecorder
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
	// chartRepoInformerFactory := informers.NewSharedInformerFactoryWithOptions(appClient, time.Second*30, informers.WithNamespace(opt.ChartRepoNamespace))

	informer := appInformerFactory.App().V1alpha1().HelmRequests()
	// repoInformer := chartRepoInformerFactory.App().V1alpha1().ChartRepos()

	controller := &Controller{
		kubeClient:   kubeClient,
		appClientSet: appClient,
		clusterConfig: clusterConfig{
			clusterNamespace:  opt.ClusterNamespace,
			clusterClient:     clusterClient,
			globalClusterName: opt.GlobalClusterName,
		},
		systemNamespace:   opt.ChartRepoNamespace,
		restConfig:        cfg,
		recorder:          mgr.GetEventRecorderFor(util.ComponentName),
		helmRequestLister: informer.Lister(),
		// chartRepoLister:    repoInformer.Lister(),
		helmRequestSynced: informer.Informer().HasSynced,
		// chartRepoSynced:    repoInformer.Informer().HasSynced,
		workQueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "HelmRequests"),
		// chartRepoWorkQueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "ChartRepos"),
		// refresh frequently
		ClusterCache: commoncache.New(1*time.Minute, 5*time.Minute),

		// only init data structures, start it later
		clusterHelmRequestListers: make(map[string]listers.HelmRequestLister),
		clusterHelmRequestSynced:  make(map[string]cache.InformerSynced),
		clusterWorkQueues:         make(map[string]workqueue.RateLimitingInterface),
		clusterClients:            make(map[string]clientset.Interface),
		clusterRecorders:          make(map[string]record.EventRecorder),
	}

	klog.Info("Setting up event handlers")
	// Set up an event handler for when HelmRequest resources change
	informer.Informer().AddEventHandler(controller.newHelmRequestHandler())
	// repoInformer.Informer().AddEventHandler(controller.newChartRepoHandler())

	klog.V(7).Infof("cluster rest config is : %+v", cfg)

	// notice that there is no need to run Start methods in a separate goroutine. (i.e. go kubeInformerFactory.Start(stopCh)
	// Start method is non-blocking and runs all registered informers in a dedicated goroutine.
	// kubeInformerFactory.Start(stopCh)
	// appInformerFactory.Start(stopCh)

	// fuck examples, this should after init controller
	appInformerFactory.Start(stopCh)
	// chartRepoInformerFactory.Start(stopCh)

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
	// defer c.chartRepoWorkQueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	klog.Info("Starting HelmRequest controller")

	// Wait for the caches to be synced before starting workers
	klog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.helmRequestSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	klog.Info("Starting workers")
	// Launch two workers to process HelmRequest resources
	for i := 0; i < 2; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
		// go wait.Until(c.runChartRepoWorker, time.Second, stopCh)
	}

	// starts other clusters
	if err := c.startAllClustersWatch(stopCh); err != nil {
		return err
	}

	klog.Info("Started workers")
	<-stopCh
	klog.Info("Shutting down workers")

	// fuck. this bug. shutdown now manually
	for _, v := range c.clusterWorkQueues {
		v.ShutDown()
	}

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

// see issue: https://github.com/kubernetes/kubernetes/issues/60845
// the origin code in sample-controller is not working... fuck
func (c *Controller) updateHelmRequestStatus(helmRequest *alpha1.HelmRequest) error {
	// NEVER modify objects from the store. It's a read-only, local cache.
	// You can use DeepCopy() to make a deep copy of original object and modify this copy
	// Or create a copy manually for better performance
	h := helm.GenUniqueHash(helmRequest)
	// Note: we have to generate the hash before the deepcopy, because somehow the deepcopy
	// can create a spec that have different hash value.
	request := helmRequest.DeepCopy()
	request.Status.LastSpecHash = h
	return c.updateHelmRequestPhase(request, alpha1.HelmRequestSynced)
}

// setPartialSyncedStatus set spec hash and partial-synced status for helm-request
//TODO: merge with updateHelmRequestStatus
func (c *Controller) setPartialSyncedStatus(helmRequest *alpha1.HelmRequest) error {
	h := helm.GenUniqueHash(helmRequest)
	request := helmRequest.DeepCopy()
	request.Status.LastSpecHash = h
	return c.updateHelmRequestPhase(request, alpha1.HelmRequestPartialSynced)
}

// setSyncFailedStatus set HelmRequest to Failed and generated a warning event
func (c *Controller) setSyncFailedStatus(helmRequest *alpha1.HelmRequest, err error) error {
	c.sendFailedSyncEvent(helmRequest, err)
	return c.updateHelmRequestPhase(helmRequest, alpha1.HelmRequestFailed)

}

func (c *Controller) setPendingStatus(helmRequest *alpha1.HelmRequest) error {
	return c.updateHelmRequestPhase(helmRequest, alpha1.HelmRequestPending)
}

// getAppClient get a kubernetes app client for the target hr(may be from global cluster or other clusters)
func (c *Controller) getAppClient(hr *alpha1.HelmRequest) clientset.Interface {
	if hr.ClusterName == "" {
		return c.appClientSet
	} else {
		return c.clusterClients[hr.ClusterName]
	}
}

// if this helmrequst deployed to a remote cluster, the release cluster will be .spec.clusterName
func (c *Controller) getAppClientForRelease(hr *alpha1.HelmRequest) clientset.Interface {
	if hr.Spec.ClusterName == "" {
		return c.getAppClient(hr)
	} else {
		if hr.Spec.ClusterName == "global" {
			return c.appClientSet
		}
		return c.clusterClients[hr.Spec.ClusterName]
	}
}

func (c *Controller) getHelmRequestLister(name string) listers.HelmRequestLister {
	if name == "" {
		return c.helmRequestLister
	} else {
		return c.clusterHelmRequestListers[name]
	}
}

func (c *Controller) getEventRecorder(hr *alpha1.HelmRequest) record.EventRecorder {
	if hr.ClusterName == "" {
		return c.recorder
	} else {
		return c.clusterRecorders[hr.ClusterName]
	}
}

// getDeployCluster returns the cluster name which the target Release lives in
func (c *Controller) getDeployCluster(hr *alpha1.HelmRequest) string {
	if hr.Spec.ClusterName != "" {
		return hr.Spec.ClusterName
	} else {
		return hr.ClusterName
	}
}
