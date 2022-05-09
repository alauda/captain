# CRD
<!-- markdown-toc start - Don't edit this section. Run M-x markdown-toc-generate-toc again -->
**Table of Contents**

- [CRD](#crd)
    - [HelmRequest](#helmrequest)
        - [spec.chart](#specchart)
        - [spec.version](#specversion)
        - [spec.releaseName](#specreleasename)
        - [spec.namespace](#specnamespace)
        - [spec.clusterName](#specclustername)
        - [spec.installToAllClusters](#specinstalltoallclusters)
        - [spec.dependencies](#specdependencies)
        - [spec.values](#specvalues)
        - [spec.valuesFrom](#specvaluesfrom)
        - [spec.source](#specsource)
    - [ChartRepo](#chartrepo)
        - [Basic Auth](#basic-auth)
        - [Type](#chartrepo-type)
    - [Chart](#chart)
    - [Release CRD](#release-crd)

<!-- markdown-toc end -->


This document will describe all the CRDs captain use and introduced.

## HelmRequest

HelmRequest is a CRD that defines how to install a helm charts. It's kind of  like the helm cli, you can set various options when install a helm charts. But with HelmRequest, it easier to persistent the configuration, and because we now have a controller, which can ensure to successfully install the chats by retrying automatically.  This can solve a lot of problems when using helm chats in helm 2, such as 

* Some pre-existing resource cause the `helm install` to fail
* CRD with `crd-install` hooks not deleted after delete a helm chart, which cause the previous problem again
* arbitrary errors which can be resolved by retry
* ... 


An example HelmRequest looks like this(after setting some defaults):



```yaml
apiVersion: app.alauda.io/v1alpha1
kind: HelmRequest
metadata:
  finalizers:
  - captain.alauda.io
  name: nginx-ingress
  namespace: default
spec:
  chart: stable/nginx-ingress
  namespace: default
  releaseName: nginx-ingress
  clusterName: ""
  dependencies:
  - nginx
  installToAllClusters: false
  values:
    env: prod
    global:
      namespace: h8
status:
  lastSpecHash: "15900028196316601597"
  notes: "<helm charts notes>"
  phase: Synced
```

### spec.chart

The chart name, format as `<repo-name>/<chart-name>`,  this is the only force required field in HelmRequest.Spec


### spec.version

The chart's version. It's optional, just likes helm cli.


### spec.releaseName
If not set, default to HelmRequest.Name. In Helm2, if you not set the release name, helm will create a random string for you(which looks like docker container names),
which is very unreasonable.

Also, since Release is namespace scoped CRD resource, it's name is not required to be global unique anymore

### spec.namespace

If not set, default to HelmRequest.Namespace


### spec.clusterName
If not set, default to "", which means this chart will be installed to the current. Otherwise, the charts will be installed to the specific cluster

### spec.installToAllClusters

Default to false, and it override `spec.clusterName`. If set to `true`, means this charts will be installed to all the clusters, not only the existing ones, 
even the ones added after this HelmRequest (informers will force rsync HelmRequest resource periodically, and captain will refresh cluster list, so just wait and see the magic happens!)


### spec.dependencies

A list of HelmRequests in the current namespace that need to be synced before this one.


### spec.values
The same format and effect as in helm's `values.yaml` file. 

### spec.valuesFrom

List of Secrets, ConfigMaps from which to take values.  If both `spec.values` and `spec.valuesFrom` is set, the `spec.values` will override.

```yaml
spec:
  # chart: ...
  valuesFrom:
  - configMapKeyRef:
      # Name of the config map, must be in the same namespace as the
      # HelmRequest 
      name: default-values  # mandatory
      # Key in the config map to get the values from
      key: values.yaml      # optional; defaults to values.yaml
      # If set to true successful retrieval of the values file is no
      # longer mandatory
      optional: false       # optional; defaults to false
  - secretKeyRef:
      # Name of the secret, must be in the same namespace as the
      # HelmRequest
      name: default-values # mandatory
      # Key in the secret to get thre values from
      key: values.yaml     # optional; defaults to values.yaml
      # If set to true successful retrieval of the values file is no
      # longer mandatory
      optional: true       # optional; defaults to false
```

Centralized configuration can be a great helper to manage multiple HelmRequest resources. 

### spec.source

Helmrequest now supports version v1 and has added new fields.

The chart's source. It's optional, this will indicate the source of the current chart, which will be an OCI or HTTP URL address.
If basic authentication is required, specify a secretname in the `spec.source.oci` or `spec.source.http`.

As shown in the following examples

```yaml
# OCI type
apiVersion: app.alauda.io/v1
kind: HelmRequest
metadata:
  name: test-oci
  namespace: default
spec:
  source:
    oci:
      repo: 192.168.26.40:60080/acp/chart-tomcat         # oci repo address
      secretRef: ociSecret      # optional, if basic authentication is required, specify a secretname here.
  values:
    namespace: default
  version: 9.2.9    # required, oci version
---
# HTTP type
apiVersion: app.alauda.io/v1
kind: HelmRequest
metadata:
  name: test-http
  namespace: default
spec:
  source:
    http:
      url: https://alauda.github.io/captain-test-charts/wordpress-lookup-11.0.13.tgz
  values: {}
```


## ChartRepo

`ChartRepo` represents a helm repository, where helm client can retrieve and upload helm charts. 
The definition is quite simple, For example, the most simplest ones is :

```yaml
apiVersion: app.alauda.io/v1alpha1
kind: ChartRepo
metadata:
  name: stable
  namespace: captain
spec:
  url: https://kubernetes-charts.storage.googleapis.com
  type: Chart
``` 

* `metadata.name`: the name of this repo
* `metadata.namespace`: Captain will only read ChartRepo resources from one namespace, 
   default to `captain`, and can be customized in helm values (`.namespace`)
* `spec.url`: the url of this repo
* `spec.type`: the type of this chartrepo. Here `Chart` means this is a normal helm chart repo. For more information, see the sections below

After created, we can use `kubectl` to checkout the repo list

```bash
root@VM-16-12-ubuntu:/home/ubuntu# kubectl get ctr -n captain
NAME     URL                                                PHASE    AGE
stable   https://kubernetes-charts.storage.googleapis.com   Synced   21m
```

The output is very similar to `helm repo list`.

### ChartRepo Type
In addition to normal helm chart repo (http server), captain also supported following use cases:
* Pull charts manifests from vcs repo, eg: git/svn
* Shipped with a built in helm chart repo, so the user can use it out-of-box

ChartRepo support following type:
* `Chart`: helm chart repo
* `Git/SVN`: refer to [Git/SVN Support](./vcs-repo.md) for more information.
* `Local`: use the chartrepo shipped with captain. In this case, `.spec.url` should be "", and captain will generate the repo url for you.

### Basic Auth
Of course ,many repos need auth support. Currently, `ChartRepo` has support basic auth by specify 
a secret resource in the spec:

```yaml
apiVersion: app.alauda.io/v1alpha1
kind: ChartRepo
metadata:
  name: new
  namespace: captain
spec:
  url: <url>
  secret:
    name: new
``` 

* `spec.secret.name`: name of the secret
* `spec.secret.namespace`: namespace of the secret, an optional field, default to the same namespace as `ChartRepo`

Then, all you need is a secret which contains `username` and `password` data:

```yaml
apiVersion: v1
data:
  password: MndiNEUxaXlkUmo3
  username: N0RPOVFvTHREeDFn
kind: Secret
metadata:
  name: new
  namespace: captain
type: Opaque
```


## Chart

From version `0.9.2`, captain add a `Chart` CRD to represents helm charts info. Here is an example chart 

```yaml
[root@ace-master-1 ~]# kubectl  get charts.app.alauda.io -n alauda-system seq.stable
NAME         VERSION   APPVERSION   AGE
seq.stable   1.0.2     5            7d
[root@ace-master-1 ~]# kubectl  get charts.app.alauda.io -n alauda-system seq.stable -o yaml
apiVersion: app.alauda.io/v1alpha1
kind: Chart
metadata:
  creationTimestamp: "2019-09-17T05:58:19Z"
  generation: 1
  labels:
    repo: stable
  name: seq.stable
  namespace: alauda-system
  resourceVersion: "49110918"
  selfLink: /apis/app.alauda.io/v1alpha1/namespaces/alauda-system/charts/seq.stable
  uid: 260e2e49-d910-11e9-b3b1-525400488bfd
spec:
  versions:
  - apiVersion: v1
    appVersion: "5"
    created: "2019-09-17T05:28:13.665649241Z"
    description: Seq is the easiest way for development teams to capture, search and
      visualize structured log events! This page will walk you through the very quick
      setup process.
    digest: 8c64cdeb44d002cec4cda2abeaf7a3fa618da34aebdb7b789512400fd3543b82
    home: https://getseq.net/
    icon: https://avatars1.githubusercontent.com/u/5898109?s=200&v=4
    keywords:
    - seq
    - structured
    - logging
    maintainers:
    - email: nblumhardt@nblumhardt.com
      name: nblumhardt
    - email: ashleymannix@live.com.au
      name: KodrAus
    - email: gertjvr@gmail.com
      name: gertjvr
    name: seq
    sources:
    - https://github.com/datalust/seq-tickets
    urls:
    - https://kubernetes-charts.storage.googleapis.com/seq-1.0.2.tgz
    version: 1.0.2
  - apiVersion: v1
    appVersion: "5"
    created: "2019-07-14T23:57:05.986935922Z"
    description: Seq is the easiest way for development teams to capture, search and
      visualize structured log events! This page will walk you through the very quick
      setup process.
    digest: 922009a14d60c136440d5ac9fdae8e5145d43b8ca3f46f24a34871b8f9cba875
    home: https://getseq.net/
    icon: https://avatars1.githubusercontent.com/u/5898109?s=200&v=4
    keywords:
    - seq
    - structured
    - logging
    maintainers:
    - email: nblumhardt@nblumhardt.com
      name: nblumhardt
    - email: ashleymannix@live.com.au
      name: KodrAus
    - email: gertjvr@gmail.com
      name: gertjvr
    name: seq
    sources:
    - https://github.com/datalust/seq-tickets
    urls:
    - https://kubernetes-charts.storage.googleapis.com/seq-1.0.1.tgz
    version: 1.0.1
```
A little explanation

* the name of this chart choose the format of `<chart-name>.<repo-name>`, because different repo can have chart with the same name
* `.spec` contains list of chart versions, which contains all the metadata a chart version have

After a [ChartRepo](https://github.com/alauda/captain/blob/master/docs/en/crds/chartrepo.md) was created ,captain will start to sync the charts this repo have and create all the charts resources,  mostly this will only take less than a minute.
For now the charts resources only serves to api usage, captain still use helm's local cache to locate chart, but this will be soon changed.

## Release CRD

Release CRD stores the release info of a helm charts, it has some differences compares to the Helm2 format:

* It's a CRD, not a special purpose ConfigMap or Secret
* No more name conflict between different namespaces
* It may lives in a different cluster with HelmRequest


Here is an example Release resource:

```yaml
apiVersion: app.alauda.io/v1alpha1
kind: Release
metadata:
  creationTimestamp: "2019-07-26T07:15:59Z"
  generation: 2
  labels:
    modifiedAt: "1564125359"
    name: nginx
    owner: helm
    status: deployed
    version: "3"
  name: nginx.v3
  namespace: h8
  resourceVersion: "27352526"
  selfLink: /apis/app.alauda.io/v1alpha1/namespaces/h8/releases/nginx.v3
  uid: 37fae69c-af75-11e9-b1c5-525400cb432d
spec:
  chartData: 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
  configData: H4sIAAAAAAAC/0SMS6oDIRBF93LHtu/DGzxqmlGWUa2FEeyysQx0aNx7MBlkcrhwD+cEM2jCYV1BEw6p1JUL6ETeOInNpSnrcdXUxOxStbdairT5NNmr5V7bA4RQmO3rJS/5bS/hozt0TiB8+98//4MxHJQ3sZ2DgHD7h0OTlK3P2gmOcSZAyBrl8Fz4HtkHxRjjGQAA//+NfINQvQAAAA==
  hooksData: H4sIAAAAAAAC/8orzckBBAAA//9P/MslBAAAAA==
  manifestData: H4sIAAAAAAAC/9xWW2/jNhP9K0T2e/j6IF/iFAgE5MGbuG3QbGqs3X0yEIypkUSEIlleHLtF/3tBSnJ0czYpmgJbvYkzc3hmeDicsyiKNuIDWUmnKcZEZEzsIyYyjcaMLRaKg0UzDuujAxR8Ix6ZSGJyLUXKsk+gNgIU+4LaMClisptuRIEWErAQbwQhAoojLg0xToNlUtRGo8DvnF/6BQ5b5CYEEgJKjR7dFrVAi2bE5LgEq+hFAfSUqwJtI5n2vP+lhC1VkUG9YxTNfztTl3wrmdZRZkerfLvplfmvymw62b5/GnVFB5cbJTUKaeCBe4taAF9rSFNGl5IzeojJnaTAvV1JbUvGUQWdW6t8qQghQia4lNrGZHY+mU5KsiosXNZ/WlpJJY/J+npZLlnQGdoyrgRrg5s++nQ2OW+iX1zMXg0fFGWQI7VSv2/t7UGhrx0kH4GDoKj/1hVqSgqUMuNnXd2g4vJQoLAv9MiaF5XCasm5p/G1S1Vdv4c31OEfKNlRhhoVZxRMTKYDx1WApfldk+6b+LzxEKszqfduVtl/vM3jjUzexiX4CyFteO6amyotC7Q5uhIhqH1zNp2cf3+xOTvlZqgGr9DNmdUOa7/qED6UEZUQ5pRKJ+z9gKwqDyg9jjTTlAlmD02OMpnXq+SPP9sGYdm8H+KV8JtjGpMbp5nIVjTHxHEmsttMyOPyYo/U+ZK0QqPybFZt8Tx/QUaLvfJZdMpZAzziIW7dha4LIVKhBg9PbkXfugPusI/ssU/oIfQsqSSX2eFnv31bHLk0tknE/9+jfZL6MSbhGCuDb5cDqVe7dg+dEN8dgAnUTbbRKxrJEbmALMg8wf0IOLgERlSMKQeo+tkAQDwZnV+Mpk0c0Fm3YBEZv4aBd4yqkbAAdfW//y9/uXm4n39arJbz68V34xMzYzu+OWe9BNadx9oozRnmJZTurNNGUW7Lmclrj374CyLy8Y1mESmNKdtflS9LFdfuPP14FLDlGBnDIwXG2FxLl+V9vwRTcNwGR4raspRRsHiVX469Gixv5WaQOs3s4VoKi3vbOWzgXD4tNdsxjhkuDAUeMmjLuxItKNgyziwbuGSJlmrw5s3v7rrLkCSDrveL9cPH2/ubh9Xi85fb60Xb5wN5enqK/HNAoo2bTGZIZrO2i3Zibn41qOOOCcWup/LystVH3OUTeskPWhYDTFOGPPmM6YCpMi7B5vHx+Rr1m1l3+6Cwd+cQRpBmwPOE2adWj4ctCdR9a9kaNIdCzVdij2Nk9bKzHQo0ZqnlFjuMUmDcaVznGk0ueRKTzrn7/X5E26uFCiUY5wjc5r/3rIFH+Wx3TIbm6BP5ab1etm3+0WTAb5DDYYVUisTPTZ06KNRMJifNxlGKxjTymXaeJFagdPYEgEZI2LdQq/cow9lfAQAA//8jyZd28xAAAA==
  name: nginx
  version: 3
status:
  Description: Upgrade complete
  deleted: null
  first_deployed: "2019-07-26T07:02:49Z"
  last_deployed: "2019-07-26T07:15:59Z"
  status: deployed
```

From the internal view, it also has some improvements compares to the old version:

* Use separated filed to store helm data
* Move some info the to `status` filed, which is more reasonable


Now you can use kubectl to get the releases and see there status directly:

```bash
[root@ake-master1 ~]# kubectl get rel --all-namespaces
NAMESPACE   NAME                          STATUS       AGE
h8          sh.helm.release.v1.nginx.v1   superseded   2d
h8          sh.helm.release.v1.nginx.v2   superseded   2d
h8          sh.helm.release.v1.nginx.v3   deployed     2d
[root@ake-master1 ~]#
```
