## HelmRequest

HelmRequest is a CRD that defines how to install a helm charts. It's kindly like the helm cli, you can set various options when install a helm charts. But with HelmRequest, it easier to persistent the configuration, and because we now have a controller, which can ensure to successfully install the chats by retrying automatically.  This can solve a lot of problems when using helm chats in helm 2, such "resource already exist" , bluh,blush….



A basic HelmRequest looks like this(after setting some defauts):



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
status:
  lastSpecHash: "15900028196316601597"
  notes: "<helm charts notes>"
  phase: Synced
```







