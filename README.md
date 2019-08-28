# Captain

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0) [![Go Report Card](https://goreportcard.com/badge/github.com/alauda/captain)](https://goreportcard.com/report/github.com/alauda/captain) [![Tag](https://img.shields.io/github/tag/alauda/captain)](https://img.shields.io/github/tag/alauda/captain)

Captain is a Helm 3 Controller

## About Helm3 

The [Helm 3 Design Proposal](https://github.com/helm/community/blob/master/helm-v3/000-helm-v3.md) has been exist for a while, and the helm 
developer group is focused on the core helm 3 development, this is the first implementation of Helm3 Controller based on the Proposal.

This project is based on the core [helm](https://github.com/helm/helm) v3 code, which promised to be act as an library. Since it's not official 
released yet (alpha stage for now), we add some little modification to help create this controller([alauda/helm](https://github.com/alauda/helm),  of course this will be unnecessary in the future.).

## Features
* HelmRequest and Release CRD, namespace based
* ChartRepo CRD
* Multi cluster support based on [https://github.com/kubernetes/cluster-registry](https://github.com/kubernetes/cluster-registry)
* Dependency check for HelmRequest (between HelmRequests)
* `valuesFrom` support, also use ConfigMap or Secret to store values
* `kubectl apply` like resource manipulation，no more resource conflict and CRD management issues


## Quick Start
Check the [Installation Guide](./docs/install.md) to learn how to install captain

Then, create a HelmRequest resource 

```yaml
kind: HelmRequest
apiVersion: app.alauda.io/v1alpha1
metadata:
  name: nginx-ingress
spec:
  chart: stable/nginx-ingress
```
After a few seconds, you have a nginx-ingress chart running

```bash
root@VM-16-12-ubuntu:~/demo# kubectl get pods
NAME                                             READY   STATUS    RESTARTS   AGE
nginx-ingress-controller-57987f445c-9rhv5        1/1     Running   0          16s
nginx-ingress-default-backend-7679dbd5c9-wkkss   1/1     Running   0          16s
root@VM-16-12-ubuntu:~/demo# kubectl get hr
NAME            CHART                  VERSION   NAMESPACE   ALLCLUSTER   PHASE    AGE
nginx-ingress   stable/nginx-ingress             default                  Synced   23s
```

For the detailed explain and advanced usage, please check the documentation below



## Documention

* [How captain works](./docs/captain.md)
* [HelmRequest CRD](./docs/helmrequest.md)
* [Release CRD](./docs/release.md)
* [ChartRepo CRD](./docs/chartrepo.md)




## Future Plans

* Release Version secret support
* Repo proxy support
* Rollback/Update kubectl plugin 
* Data migration tool from Helm2 release to Helm3 Release


## Related Project

* [flux](https://github.com/fluxcd/flux): flux have a similar controller based on Helm2



