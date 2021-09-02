## Chart

从版本`0.9.2` 开始，`captain`添加了一个`Chart` 的自定义资源，来表示helm chart的信息。 下面是一个简单的`Chart`示例

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
一些解释说明

* 不同仓库中的chart可以使用相同的名称进行定义，所以chart的名称格式为`<chart-name>.<repo-name>`
* `.spec`包含了chart的版本列表，其中包含了chart版本的所有元数据信息

创建[ChartRepo](https://github.com/alauda/captain/blob/master/docs/en/crds/chartrepo.md)后，`captain`将会开始同步仓库中的charts并进行资源创建，charts同步、创建的过程，多数情况下时间不会超过一分钟。
目前，charts资源仅用户API使用，`captain`依然还是使用helm的本地缓存来定位本地chart，这种实现方式将来会进行调整、优化。