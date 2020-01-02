# Installation Guide

```bash
# or choose a namespace you like
kubectl create ns captain-system
kubectl create clusterrolebinding captain --serviceaccount=captain-system:default --clusterrole=cluster-admin
kubectl apply -n captain-system -f https://raw.githubusercontent.com/alauda/captain/remove-cert-manager/artifacts/all/deploy.yaml
```

