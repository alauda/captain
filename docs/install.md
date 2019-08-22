# Installation Guide


## install cert-manger
The cert-manager [docs](https://docs.cert-manager.io/en/latest/getting-started/install/kubernetes.html) have details steps about how to install cert-manager,
here is the code snippet i copied from it 

```bash
# Install the CustomResourceDefinition resources separately
kubectl apply -f https://raw.githubusercontent.com/jetstack/cert-manager/release-0.8/deploy/manifests/00-crds.yaml

# Create the namespace for cert-manager
kubectl create namespace cert-manager

# Label the cert-manager namespace to disable resource validation
kubectl label namespace cert-manager certmanager.k8s.io/disable-validation=true

# Add the Jetstack Helm repository
helm repo add jetstack https://charts.jetstack.io

# Update your local Helm chart repository cache
helm repo update

# Install the cert-manager Helm chart
helm install \
  --name cert-manager \
  --namespace cert-manager \
  --version v0.8.1 \
  jetstack/cert-manager
```


## install captain
``` bash
helm repo add alauda https://alauda.github.io/charts
kubectl create namespace captain # or choose a namespace you likes， just remember to update the args below
helm install  --name=captain  --namespace=captain --set namespace=captain alauda/captain
```
