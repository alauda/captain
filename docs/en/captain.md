# How captain works

Captain is regular kubernetes controller, it watch specific resource (HelmRequest in this context), process them, and sync their states. Besides that, there are some custom components related to helm need a little explain



## Helm Repo

Since Helm3 code is under active development, it not totally ready to act as an library. So some of the helm function still need the old fashion way. Helm2 use local files to store repo information and index. 
In Captain, it remains the same way. But it can read third-party repo from `ChartRepo` CRD, which users can read and write directly using `kubectl`,
Here is the default ChartRepo which captain will install automatically when start: 


```yaml
apiVersion: app.alauda.io/v1alpha1
kind: ChartRepo
metadata:
  creationTimestamp: "2019-08-09T08:04:16Z"
  generation: 2
  name: stable
  namespace: captain
  resourceVersion: "7253523"
  selfLink: /apis/app.alauda.io/v1alpha1/namespaces/captain/chartrepos/stable
  uid: 48515f13-ba7c-11e9-98c3-5254004f2ad2
spec:
  url: https://kubernetes-charts.storage.googleapis.com
status:
  phase: Synced
```
For detaild information about ChartRepo, please checkout [ChartRepo CRD](./chartrepo.md)

## Clusters

Captain has built in support for multi-cluster, based on the kubernetes [cluster-registry](https://github.com/kubernetes/cluster-registry) project, which means you can not only install a Helm charts to the local cluster, you can also install the charts to any other cluster you specified. Besides that, there is an alternative option which allow you to install one charts to all the clusters.
This can be very convenient at production environment which always have many clusters and required to install some base component to all the clusters. 






