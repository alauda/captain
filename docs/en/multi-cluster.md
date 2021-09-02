# Multi Cluster

Multi is an optional but import feature for captain.It's based on the [cluster-registry](https://github.com/kubernetes/cluster-registry), which introduce a CRD called `Cluster`. So If you have `Cluster` in you current kubernetes env(where captain deployed to), captain will watch the clusters and sync HelmRequest for them.

## How Captain Discover Clusters
`Captain` have a command line args called `--cluster-namespace`, which specified the namespace captain will look up into to find `Cluster` resources.

## Deploy a HelmRequest to a Remote Cluster
If `.spec.clusterName` is not empty, and it's a valid cluster name, captain will deploy this HelmRequest to the target cluster. For example:

```yaml
apiVersion: app.alauda.io/v1alpha1
kind: HelmRequest
metadata:
  name: nginx-ingress
  namespace: default
spec:
  chart: stable/nginx-ingress
  namespace: default
  clusterName: "cluster1"
  values:
    env: prod
```

In the above example, `Captain` will deploy this HelmRequest to `cluster1`. In this scenario, the generated `Release` and resources will exist in `cluster`, but `HelmRequest` exist in current cluster.


## Deploy a HelmRequest to All Clusters
If `.spec.installToAllClusters` is `true`, Captain will deploy the HelmRequest to all the clusters it knows.

```yaml
apiVersion: app.alauda.io/v1alpha1
kind: HelmRequest
metadata:
  name: nginx-ingress
  namespace: default
spec:
  chart: stable/nginx-ingress
  namespace: default
  installToAllClusters: true
  values:
    env: prod
```
