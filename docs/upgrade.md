# Captain upgrade

This document will describe how to upgrade captain itself.


## Helm install -> YAML install
If you have use helm to installed old version of captain, Please following the following steps to uninstall it first and 
use the latest yaml intall

### Uninstall captain installed by helm



```bash
kubectl delete mutatingwebhookconfigurations.admissionregistration.k8s.io captain
kubectl delete validatingwebhookconfigurations.admissionregistration.k8s.io captain
kubectl delete svc -n captain-system captain
kubectl delete issuer captain-selfsigned-issuer -n captain-system
kubectl delete cert captain-serving-cert -n captain-system
kubectl delete deploy captain -n captain-system
kubectl delete secret captain-webhook-cert -n captain-system
# not found error can be ignored
kubectl get cm -n kube-system |grep ^captain | awk '{print $1}' | xargs kubectl delete cm -n kube-system
 
```

### Install captain use yaml 

Please refer to [Install Guide](./install.md)



## Normally Upgrade

Just edit captain's deployment to update image tag


## Upgrade from ChartRepo v1alpha1 -> v1beta1
If you already have v1alpha1 ChartRepo exist, captain will handler it correctly. But, if you want to use the latest 
v1beta1 version, you can manually edit the ChartRepo resource to add type to it. For all of the v1alpha1 ChartRepo, 
it should be `Chart`. For more details and examples, please refer to [Git/SVN as ChartRepo](./vcs-repo.md)