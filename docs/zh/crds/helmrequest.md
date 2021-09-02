## HelmRequest

`HelmRequest`是一个CRD，它定义了如何安装一个 helm charts。 它有点像 `helm cli`，你可以在安装 helm charts 时设置各种选项。 
但是使用 HelmRequest，更容易持久化配置，并且因为我们现在有一个控制器，它可以通过自动重试来确保成功安装chart。 
这样可以解决很多在 helm 2 中使用 helm chats 的问题，比如：

* 一些预先存在的资源导致`helm install`失败
* 删除一个 helm chart 后没有删除带有 `crd-install` hook 的 CRD，这又导致了之前的问题
* 可以通过重试解决的任意错误
* ... 

HelmRequest 示例如下（设置一些默认值后）：

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

`chart`的名称，格式为 `<repo-name>/<chart-name>`，这是 `HelmRequest.Spec` 中唯一的强制要求字段

## spec.version

`chart`的版本。 它是可选的，就像 helm cli 一样。

## spec.releaseName

如果未设置，则默认为 `HelmRequest.Name`。 在 Helm2 中，如果你没有设置release名称，helm 会为你创建一个随机字符串（看起来像 docker 容器名称），这是非常不合理的。

此外，由于 Release 是命名空间范围的 CRD 资源，因此它的名称不再需要是全局唯一的。

## spec.namespace

如果没有设置，默认为 `HelmRequest.Namespace`


## spec.clusterName

如果没有设置，默认为""，表示这个chart将安装到当前集群。 否则，图表将安装到特定集群。

## spec.installToAllClusters

Default to false, and it override `spec.clusterName`. If set to `true`, means this charts will be installed to all the clusters, not only the existing ones, 
even the ones added after this HelmRequest (informers will force rsync HelmRequest resource periodically, and captain will refresh cluster list, so just wait and see the magic happens!)

默认为 false，它会覆盖 `spec.clusterName`。 如果设置为`true`，则表示此图表将安装到所有集群，而不仅仅是现有集群，
甚至在这个 HelmRequest 之后添加的那些（informers 会定期强制 rsync HelmRequest 资源，并且`captain`会刷新集群列表，所以等着看奇迹发生吧！）

## spec.dependencies

当前命名空间中需要在此之前同步的 HelmRequest 列表。

## spec.values

与 helm 的 `values.yaml` 文件中的格式和效果相同。

## spec.valuesFrom

Secrets资源列表，ConfigMaps从其中取值。 如果`spec.values` 和`spec.valuesFrom` 都被设置，`spec.values` 将被覆盖。

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

集中配置可以成为管理多个 HelmRequest 资源的好帮手。