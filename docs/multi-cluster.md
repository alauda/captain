# Multi Cluster

Multi is an optional but import feature for captain.It's based on the [cluster-registry](https://github.com/kubernetes/cluster-registry), which introduce a CRD called `Cluster`. So If you have `Cluster` in you current kubernetes env(where captain deployed to), captain will watch the clusters and sync HelmRequest for them.
