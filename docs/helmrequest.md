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

## spec.chart

The chart name, format as `<repo-name>/<chart-name>`,  this is the only force required field in HelmRequest.Spec


## spec.version

The chart's version. It's optional, just likes helm cli.


## spec.releaseName
If not set, default to HelmRequest.Name. In Helm2, if you not set the release name, helm will create a random string for you(which looks like docker container names),
which is very unreasonable.

Also, since Release is namespace scoped CRD resource, it's name is not required to be global unique anymore

## spec.namespace

If not set, default to HelmRequest.Namespace


## spec.clusterName
If not set, default to "", which means this chart will be installed to the current. Otherwise, the charts will be installed to the specific cluster

## spec.installToAllClusters

Default to false, and it override `spec.clusterName`. If set to `true`, means this charts will be installed to all the clusters, not only the existing ones, 
even the ones added after this HelmRequest (informers will force rsync HelmRequest resource periodically, and captain will refresh cluster list, so just wait and see the magic happens!)


## spec.dependencies

A list of HelmRequests in the current namespace that need to be synced before this one.


## spec.values
The same format and effect as in helm's `values.yaml` file. 

## spec.valuesFrom

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









