package controller

import (
	"fmt"
	"github.com/alauda/captain/pkg/cluster"
	"os"
	"strings"
	"time"

	"github.com/alauda/captain/pkg/util"
	alpha1 "github.com/alauda/helm-crds/pkg/apis/app/v1alpha1"
	clientset "github.com/alauda/helm-crds/pkg/client/clientset/versioned"
	hrScheme "github.com/alauda/helm-crds/pkg/client/clientset/versioned/scheme"
	informers "github.com/alauda/helm-crds/pkg/client/informers/externalversions"
	corev1 "k8s.io/api/core/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
)

// initClusterWatches init cluster watch state, and start all the factories
func (c *Controller) initClusterWatches(stopCh <-chan struct{}) error {
	clusters, err := c.getAllClusters()
	if err != nil {
		return err
	}

	// for event recording?
	hrScheme.AddToScheme(scheme.Scheme)

	for _, cluster := range clusters {
		// TODO: fix it
		if cluster.Name == "global" {
			continue
		}

		if err := c.initWatchForCluster(stopCh, cluster); err != nil {
			continue
		}
	}
	return nil
}

// initWatchForCluster init watch structs for a single cluster
func (c *Controller) initWatchForCluster(stopCh <-chan struct{}, cluster *cluster.Info) error {
	cfg := cluster.ToRestConfig()
	client, err := clientset.NewForConfig(cfg)
	if err != nil {
		klog.Warningf("init client for cluster %s error: %s", cluster.Name, err.Error())
		return err
	}

	coreClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Warningf("init core client for cluster %s error: %s", cluster.Name, err.Error())
		return err
	}

	if err := util.InstallHelmRequestCRD(cfg); err != nil {
		klog.Warningf("install helmrequest crd for cluster %s error: %s", cluster.Name, err.Error())
		return err
	}

	informerFactory := informers.NewSharedInformerFactory(client, defaultResyncDuration)
	informer := informerFactory.App().V1alpha1().HelmRequests()

	c.clusterHelmRequestListers[cluster.Name] = informer.Lister()
	c.clusterHelmRequestSynced[cluster.Name] = informer.Informer().HasSynced
	c.clusterWorkQueues[cluster.Name] = workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), cluster.Name)
	c.clusterClients[cluster.Name] = client
	c.clusterRecorders[cluster.Name] = c.createEventRecorder(cluster.Name, coreClient)

	// add event handler
	informer.Informer().AddEventHandler(c.newClusterHelmRequestHandler(cluster.Name))

	informerFactory.Start(stopCh)

	klog.Info("init watch config for cluster: ", cluster.Name)
	return nil
}

func (c *Controller) CleanupClusterWatch(name string) {
	c.clusterHelmRequestListers[name] = nil
	c.clusterHelmRequestSynced[name] = nil
	c.clusterWorkQueues[name] = nil
	c.clusterClients[name] = nil
	c.clusterRecorders[name] = nil
}

// IsClusterWatchStarted check if a cluster watch has been started
func (c *Controller) IsClusterWatchStarted(name string) bool {
	return c.clusterClients[name] != nil
}

// restartClusterWatch will restart the failed cluster watches. In this situation, all the hr will be failed at get release client ,
// so we will trigger from there and try to re-init the watch and restart it
func (c *Controller) restartClusterWatch(cluster *cluster.Info) error {

	if err := c.initWatchForCluster(c.stopCh, cluster); err != nil {
		return err
	}

	if err := c.startClusterWatch(cluster.Name, c.stopCh); err != nil {
		return err
	}

	klog.Info("restart watch for cluster done", cluster.Name)
	return nil

}

// createEventRecorder create event recoder for a cluster
// create the recoder manually is easier to user the method provides by controller-runtime.Manager. Maybe?
// TODO: change all args of cluster to cluster (from `name`)
func (c *Controller) createEventRecorder(cluster string, client kubernetes.Interface) record.EventRecorder {
	klog.Info("Creating event broadcaster for cluster: ", cluster)
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(
		&v1.EventSinkImpl{
			Interface: client.CoreV1().Events(""),
		},
	)
	// In case two global exist, and captain caches access info for the same cluster, this can help us to
	// find which captain create the event. Default to ""
	hostIP := os.Getenv("MY_POD_IP")
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: cluster, Host: hostIP})
	return recorder
}

// start!
func (c *Controller) startAllClustersWatch(stopCh <-chan struct{}) error {
	if err := c.initClusterWatches(stopCh); err != nil {
		return err
	}

	klog.Info("init cluster helmrequests watches done")

	for k := range c.clusterClients {
		if err := c.startClusterWatch(k, stopCh); err != nil {
			return err
		}
	}

	return nil

}

// startClusterWatch start watches for other clusters...vision is back!
func (c *Controller) startClusterWatch(name string, stopCh <-chan struct{}) error {
	// defer utilruntime.HandleCrash()
	// defer c.clusterWorkQueues[name].ShutDown()

	// Start the informer factories to begin populating the informer caches
	klog.Info("Starting HelmRequest controller for cluster: ", name)

	// Wait for the caches to be synced before starting workers
	klog.Info("Waiting for informer caches to sync: ", name)
	if ok := cache.WaitForCacheSync(stopCh, c.clusterHelmRequestSynced[name]); !ok {
		return fmt.Errorf("failed to wait for caches to sync: %s", name)
	}

	klog.Info("Starting workers for clusters: ", name)
	// Launch two workers to process HelmRequest resources
	f := func() {
		c.runClusterWorker(name)
	}

	for i := 0; i < 2; i++ {
		go wait.Until(f, time.Second, stopCh)
	}

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workQueue.
func (c *Controller) runClusterWorker(name string) {
	for c.processNextClusterWorkItem(name) {
		//TODO: delete
		//klog.Infof("processing for :", name)
	}
}

// processNextWorkItem will read a single work item off the workQueue and
// attempt to process it, by calling the syncHandler.
func (c *Controller) processNextClusterWorkItem(name string) bool {
	queue := c.clusterWorkQueues[name]

	obj, shutdown := queue.Get()

	if shutdown {
		klog.Warning("work queue has closed: ", name)
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
		defer queue.Done(obj)
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
			queue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workQueue but got %#v", obj))
			return nil
		}
		// Start the syncHandler, passing it the namespace/name string of the
		// HelmRequest resource to be synced.
		if err := c.syncHandler(key); err != nil {
			// Put the item back on the workQueue to handle any transient errors.
			queue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		queue.Forget(obj)
		klog.Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

// enqueueHelmRequest takes a HelmRequest resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than HelmRequest.
func (c *Controller) enqueueClusterHelmRequest(obj interface{}, name string) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}

	key = fmt.Sprintf("%s/%s", name, key)
	klog.Infof("enqueue helmrequest: %s", key)

	c.clusterWorkQueues[name].Add(key)
}

func clusterKey(key, name string) string {
	return fmt.Sprintf("%s/%s", name, key)
}

// example: global/default/a -> global,default/a
// note: this only works for namespaced resource,in this case , HelmRequest
func splitClusterKey(key string) (string, string) {
	ss := strings.Split(key, "/")
	if len(ss) == 2 {
		return "", key
	}
	cluster := ss[0]
	k := key[len(cluster)+1:]
	return cluster, k
}

func (c *Controller) deleteClusterHelmRequestHandler(obj interface{}, name string) {
	var err error
	var key string
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}

	hr := obj.(*alpha1.HelmRequest)
	klog.Infof("receive delete event, cluster %s, : %+v", name, hr)

	hr = hr.DeepCopy()
	hr.ClusterName = name

	outdated, err := c.isOldEvent(name, hr)
	if err != nil {
		c.sendFailedDeleteEvent(hr, err)
		utilruntime.HandleError(err)
		c.clusterWorkQueues[name].AddRateLimited(clusterKey(key, name))
		return
	}

	if outdated {
		return
	}

	err = c.deleteHelmRequest(hr)
	if err != nil {
		c.sendFailedDeleteEvent(hr, err)
		utilruntime.HandleError(err)
		c.clusterWorkQueues[name].AddRateLimited(clusterKey(key, name))
	} else {
		c.getEventRecorder(hr).Event(hr, corev1.EventTypeNormal, SuccessfulDelete,
			fmt.Sprintf("Deleted HelmRequest: %s", hr.GetName()))
	}
}
