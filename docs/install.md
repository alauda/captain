# Installation Guide

```bash
kubectl create ns captain-system
kubectl create clusterrolebinding captain --serviceaccount=captain-system:default --clusterrole=cluster-admin
kubectl apply -n captain-system -f https://raw.githubusercontent.com/alauda/captain/master/artifacts/all/deploy.yaml
```

