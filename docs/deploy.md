# How to deploy captain

## Run local binary

Command is 

```bash
./captain -kubeconfig=<kubeconfig path> --enable-leader-election=false
```
Explain 

* kubeconfig: the kube config path. captain will connect to the cluster in the default context.
* --leader-election=false: disable leader election. Leader election only enabled in in-cluster environment

Run `./captain --help` to see more options

## Run in kubernetes

```bash
kubectl apply -f ${CAPTAIN}/artifacts/deploy/webhook/
kubectl apply -f ${CAPTAIN}/artifacts/deploy/
```

There are a lot of resources to be created....Mainly because our webhook have to be run on HTTPS. Details explain

* ConfigMap: helm repo data, same format as helm's repo file
* Deployment: the main process, contains controller and webhook
* Service: for webhook service
* ValidatingWebhookConfiguration: webhook configuration. The `caBundle` field's data 
* Issuer: ca issuer
* Certificate: define service name and which secret the cert data will be stored in
