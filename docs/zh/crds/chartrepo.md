## ChartRepo

`ChartRepo`代表一个helm仓库，helm客户端可以上传、下载helm charts。
`ChartRepo`的定义非常简单，例如，最简单的一种方式：

```yaml
apiVersion: app.alauda.io/v1alpha1
kind: ChartRepo
metadata:
  name: stable
  namespace: captain
spec:
  url: https://kubernetes-charts.storage.googleapis.com
``` 

* `metadata.name`: helm仓库的名称
* `metadata.namespace`: `Captain`只会从一个命名空间中读取`ChartRepo`资源，默认为`captain`，用户也可以自定义。
* `spec.url`: helm仓库的地址

`ChartRepo`创建成功后，我们可以使用`kubectl`命令来查询仓库列表

```bash
root@VM-16-12-ubuntu:/home/ubuntu# kubectl get ctr -n captain
NAME     URL                                                PHASE    AGE
stable   https://kubernetes-charts.storage.googleapis.com   Synced   21m
```

上面的仓库列表输出，与执行`helm repo list`命令输出非常相似

### 基础认证

一些情况下，很多仓库需要认证支持。目前，`ChartRepo`已经支持了认证，可以通过指定一个`secret`资源（spec.secret）来实现：

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

* `spec.secret.name`: `secret`资源名称
* `spec.secret.namespace`: `secret`资源所属命名空间，为可选字段，默认情况下与`ChartRepo`在同一命名空间下

`secret`资源定义需要包含`username`、`password`字段信息：

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