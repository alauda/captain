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
``` 

* `metadata.name`: the name of this repo
* `metadata.namespace`: Captain will only read ChartRepo resources from one namespace, 
   default to `captain`, and can be customized in helm values (`.namespace`)
* `spec.url`: the url of this repo

After created, we can use `kubectl` to checkout the repo list

```bash
root@VM-16-12-ubuntu:/home/ubuntu# kubectl get ctr -n captain
NAME     URL                                                PHASE    AGE
stable   https://kubernetes-charts.storage.googleapis.com   Synced   21m
```

The output is very similar to `helm repo list`.

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