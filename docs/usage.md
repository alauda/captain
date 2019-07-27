## Basic Usage

Let's find out how simple we can install a helm chart with captain


First, create a new demo namespace

```bash
kubectl create ns demo
```

Then create a HelmRequest resource

```yaml
kind: HelmRequest
apiVersion: app.alauda.io/v1alpha1
metadata:
  name: nginx-ingress
  namespace: demo
spec:
  chart: stable/nginx-ingress
``` 

After a few seconds, you should see the HelmRequest has been synced and the chart has been installed

```bash
root@VM-16-12-ubuntu:~/demo# kubectl get hr -n demo
NAME            CHART                  VERSION   NAMESPACE   ALLCLUSTER   PHASE    AGE
nginx-ingress   stable/nginx-ingress             demo                     Synced   77s
root@VM-16-12-ubuntu:~/demo# kubectl get pods -n demo
NAME                                             READY   STATUS    RESTARTS   AGE
nginx-ingress-controller-78db9fb87-5hmjl         1/1     Running   0          80s
nginx-ingress-default-backend-7679dbd5c9-8mvpn   1/1     Running   0          79s
```


