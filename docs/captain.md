# How captain works

Captain is regular kubernetes controller, it watch some specific resource (HelmRequest in this context), process them, and sync their states. Besides that, there are some custom componets related to helm need a little explain



## Repo

Since Helm3 code is under active development, it not totally ready to act as an library. So some of the helm function still need the old fahsion. Helm 2 use local files to store repo information and index. In Captain, it remains the same way. But it mount the repo file from a kubernetes config(ConfigMap, Secret). For example,  in our depoyment, the captain charts contains a ConfigMap looks like this:



```yaml
apiVersion: v1
data:
  repositories.yaml: |
    apiVersion: v1
    generated: 2019-06-19T17:26:28.715546186+08:00
    repositories:
    - caFile: ""
      cache: /root/.helm/repository/cache/stable-index.yaml
      certFile: ""
      keyFile: ""
      name: stable
      password: ""
      url: https://kubernetes-charts.storage.googleapis.com
      username: ""
kind: ConfigMap
metadata:
  creationTimestamp: "2019-07-27T07:48:36Z"
  name: captain
  namespace: captain
  resourceVersion: "4886209"
  selfLink: /api/v1/namespaces/captain/configmaps/captain
  uid: f100c3eb-b042-11e9-bf4f-5254004f2ad2
```



You can see the content format is the same as helm client. If you need to add custom repo to captain, you will have to edit this ConfigMap and restart the Captain deployment. Captain will perioldyy update the repo index to keep the local cache update to date.



The future plan is to use a CRD to define repo info





## Clusters

Captain has built in support for multi-cluster, based on the kubernetes [cluster-registry](https://github.com/kubernetes/cluster-registry) project, which means you can not only install a Helm charts to the local cluster, you can also install the charts to any other cluster you specificed. Besides that, there is an alternative options which you can install one charts to all the clusters.







