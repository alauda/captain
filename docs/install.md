# Installation Guide


## Install

```bash
kubectl create ns captain-system
kubectl create clusterrolebinding captain --serviceaccount=captain-system:default --clusterrole=cluster-admin
kubectl apply -n captain-system -f https://raw.githubusercontent.com/alauda/captain/master/artifacts/all/deploy.yaml
```

### Use master build image
Captain use github action to auto build image for master branch. If you want to use the latest image, please update the yaml file, and replace captain image to 

```bash
alaudapublic/captain:latest
```




## Uninstall
```bash
kubectl delete -n  captain-system -f https://raw.githubusercontent.com/alauda/captain/master/artifacts/all/deploy.yaml
kubectl delete clusterrolebinding captain
kubectl delete ns captain-system
```
