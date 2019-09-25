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

After a [ChartRepo](https://github.com/alauda/captain/blob/master/docs/chartrepo.md) was created ,captain will start to sync the charts this repo have and create all the charts resources,  mostly this will only take less than a minute.
For now the charts resources only serves to api usage, captain still use helm's local cache to locate chart, but this will be soon changed.
