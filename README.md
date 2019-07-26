# Captain

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0) [![Go Report Card](https://goreportcard.com/badge/github.com/alauda/captain)](https://goreportcard.com/report/github.com/alauda/captain)

Captain is a Helm 3 Controller

## About Helm3 

The [Helm 3 Design Proposal](https://github.com/helm/community/blob/master/helm-v3/000-helm-v3.md) has been exist for a while, and the helm 
developer group is focused on the core helm 3 development, this is the first implementation of Helm3 Controller based on the Proposal.

This project is based on the core [helm](https://github.com/helm/helm) code, which promised to be act as an library. Since it's not official 
released yet, we add some little medication to help create this controller. Of course this will be unnecessary in the future. 

## Features
* HelmRequest and Release CRD, namespace based
* Multi cluster support
* Dependency check for HelmRequest
* `valuesFrom` support, also use ConfigMap or Secret to store values
* `kubectl apply` like resource manipulation，no more resource conflict 


## Quick Install
Check the [Installation Guide](./docs/install.md) to learn how to install captain



## TODO

* Release Version secret support
* Repo CRD


## Related Project

* [flux](https://github.com/fluxcd/flux)



