# Installation Guide


``` bash
helm repo add alauda https://alauda.github.io/charts
helm install  --name cert-manager --namespace cert-manager --version v0.6.6 stable/cert-manager
helm install  --name=captain --version=v2.2-b.2 --namespace=captain --set namespace=captain alauda/captain
```
