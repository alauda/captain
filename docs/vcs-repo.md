# Git/SVN as ChartRepo

For helm repo, it's normally a http based server, hosts binrary format of charts. For many users, it's not so easy to host or create a helm repo for there own. In captain, user can just provider theire vcs url, which contains the source files of one or multiple helm charts , and captain will generate a helm repo automatilly.

## ChartRepo example

```yaml
apiVersion: app.alauda.io/v1beta1
kind: ChartRepo
metadata:
  name: svn
  namespace: captain-system
spec:
  secret:
    name: svn
  source:
    path: /
    url: http://example.svn.org/svn/repo1
  type: SVN
```

It's pretty clear in this example, we provider a svn repo `http://example.svn.org/svn/repo1`, it's access creaditions is stored in the `svn` secret, and in the `/` path, it contains some charts source files. After we provider this ChartRepo to Captain, Captain will sync it and generate a http chart repo for it, it's address can be found in the synced yaml of this resource

```yaml
apiVersion: app.alauda.io/v1beta1
kind: ChartRepo
metadata:
  name: svn
  namespace: captain-system
spec:
  secret:
    name: svn
  source:
    path: /
    url: http://example.svn.org/svn/repo1
  url: http://captain-chartmuseum:8080/svn18
  type: SVN
status:
  phase: Synced
```

In the `.spec.url` field, we can see the helm chart repo url Captain generated for us. It's a in-cluster adress which can be accessed from captain(of course other pods in the same namespace).


## Path structure

Captain support two path structures:

1. the path is a dictory contains a charts' source files
2. the path contains multiple charts.

