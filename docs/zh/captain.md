# captain工作原理

Captain是kubernetes的常规控制器，它监听特定的资源（这里指的是HelmRequest），处理它们，并同步它们的状态。除此之外，还有一些与helm相关的自定义组件需要稍加说明



## Helm Repo

由于Helm3代码在持续开发中，它还没有完全准备好充当依赖。因此，一些helm功能仍然需要旧的方式。Helm2使用本地文件存储仓库和索引信息。在captian中，情况依然如此。但它可以从`ChartRepo` CRD读取第三方仓库，用户可以使用`kubectl`直接读写，以下是captain启动时自动安装的默认ChartRepo：


```yaml
apiVersion: app.alauda.io/v1alpha1
kind: ChartRepo
metadata:
  creationTimestamp: "2019-08-09T08:04:16Z"
  generation: 2
  name: stable
  namespace: captain
  resourceVersion: "7253523"
  selfLink: /apis/app.alauda.io/v1alpha1/namespaces/captain/chartrepos/stable
  uid: 48515f13-ba7c-11e9-98c3-5254004f2ad2
spec:
  url: https://kubernetes-charts.storage.googleapis.com
status:
  phase: Synced
```
有关ChartRepo的详细信息，请查看 [ChartRepo CRD](./chartrepo.md)

## Clusters

Captain基于kubernetes [cluster-registry](https://github.com/kubernetes/cluster-registry) 项目构建了对多集群的支持，这意味着您不仅可以将Helm charts安装到本地集群，还可以将这些charts安装到指定的任何其他集群。除此之外，还有一个可选项，允许您将一个charts安装到所有集群。这在生产环境中非常方便，因为生产环境总是有许多集群，并且需要为所有集群安装一些基本组件。





