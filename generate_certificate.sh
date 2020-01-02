#!/usr/bin/env sh

set -e

usage() {
  cat <<EOF
Generate certificate suitable for use with any Kubernetes Mutating Webhook.
This script uses k8s' CertificateSigningRequest API to a generate a
certificate signed by k8s CA suitable for use with any Kubernetes Mutating Webhook service pod.
This requires permissions to create and approve CSR. See
https://kubernetes.io/docs/tasks/tls/managing-tls-in-a-cluster for
detailed explantion and additional instructions.
The server key/cert k8s CA cert are stored in a k8s secret.
usage: ${0} [OPTIONS]
The following flags are required.
    --service                   Service name of webhook.
    --vwebhook                  Validating webhook config name.
    --mwebhook                  Mutating webhook config name.
    --namespace                 Namespace where webhook service and secret reside.
    --secret                    Secret name for CA certificate and server certificate/key pair.
EOF
  exit 1
}

while [ $# -gt 0 ]; do
  case ${1} in
      --service)
          service="$2"
          shift
          ;;
      --vwebhook)
          vwebhook="$2"
          shift
          ;;
      --mwebhook)
	  mwebhook="$2"
	  shift
	  ;;
      --secret)
          secret="$2"
          shift
          ;;
      --namespace)
          namespace="$2"
          shift
          ;;
      *)
          usage
          ;;
  esac
  shift
done

[ -z "${service}" ] && echo "ERROR: --service flag is required" && exit 1
[ -z "${vwebhook}" ] && echo "ERROR: --validate-webhook flag is required" && exit 1
[ -z "${mwebhook}" ] && echo "ERROR: --mutate-webhook flag is required" && exit 1
[ -z "${secret}" ] && echo "ERROR: --secret flag is required" && exit 1
[ -z "${namespace}" ] && echo "ERROR: --namespace flag is required" && exit 1

if [ ! -x "$(command -v openssl)" ]; then
  echo "openssl not found"
  exit 1
fi

csrName=${service}.${namespace}
tmpdir=$(mktemp -d)
echo "creating certs in tmpdir ${tmpdir} "

cat <<EOF >> "${tmpdir}/csr.conf"
[req]
req_extensions = v3_req
distinguished_name = req_distinguished_name
[req_distinguished_name]
[ v3_req ]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names
[alt_names]
DNS.1 = ${service}
DNS.2 = ${service}.${namespace}
DNS.3 = ${service}.${namespace}.svc
EOF

openssl genrsa -out "${tmpdir}/server-key.pem" 2048
openssl req -new -key "${tmpdir}/server-key.pem" -subj "/CN=${service}.${namespace}.svc" -out "${tmpdir}/server.csr" -config "${tmpdir}/csr.conf"


set +e

echo "secret namespace is ${namespace}"

kubectl create secret generic ${secret} -n ${namespace}

## check if secret synced, if synced exist , or annoate it and continue process...
if kubectl get secret -n ${namespace} ${secret} -o yaml |grep "cert-created"; then
    echo "Secret synced, exit"
    exit 0
fi



num=$(shuf -i 1-1000 -n 1)

kubectl annotate secret -n ${namespace} ${secret} init-pod-rand=$num --overwrite

## acciure lock and do the work.
if kubectl get secret -n ${namespace} ${secret} -o yaml |grep init-pod-rand |grep $num
then
    echo "I'm continue to create secret"
else
    while true; do
	if kubectl get secret -n ${namespace} ${secret} -o yaml | grep "cert-created" ; then
	    echo "Cert secret synced, exit"
	    exit 0
	fi
	echo "wait for cert secret to be ready..."
	sleep 1
    done
fi




 set +e
# clean-up any previously created CSR for our service. Ignore errors if not present.
 if kubectl delete csr "${csrName}"; then 
    echo "WARN: Previous CSR was found and removed."
fi
 set -e

# create server cert/key CSR and send it to k8s api
cat <<EOF | kubectl create -f -
apiVersion: certificates.k8s.io/v1beta1
kind: CertificateSigningRequest
metadata:
  name: ${csrName}
spec:
  groups:
  - system:authenticated
  request: $(base64 < "${tmpdir}/server.csr" | tr -d '\n')
  usages:
  - digital signature
  - key encipherment
  - server auth
EOF

set +e
# verify CSR has been created
while true; do
  if kubectl get csr "${csrName}"; then
      break
  fi
done
set -e

# approve and fetch the signed certificate
kubectl certificate approve "${csrName}"

set +e
# verify certificate has been signed
i=1
while [ "$i" -ne 5 ]
do
  serverCert=$(kubectl get csr "${csrName}" -o jsonpath='{.status.certificate}')
  if [ "${serverCert}" != '' ]; then
      break
  fi
  sleep 5
  i=$((i + 1))
done

set -e
if [ "${serverCert}" = '' ]; then
  echo "ERROR: After approving csr ${csrName}, the signed certificate did not appear on the resource. Giving up after 10 attempts." >&2
  exit 1
fi

echo "${serverCert}" | openssl base64 -d -A -out "${tmpdir}/server-cert.pem"


# cp ${tmpdir}/server-key.pem /tmp/k8s-webhook-server/serving-certs/tls.key
# cp ${tmpdir}/server-cert.pem /tmp/k8s-webhook-server/serving-certs/tls.crt

# create the secret with CA cert and server cert/key
# kubectl create secret tls "${secret}" \
#     --key="${tmpdir}/server-key.pem" \
#     --cert="${tmpdir}/server-cert.pem" \
#     --dry-run -o yaml |
#    kubectl -n "${namespace}" apply -f -




caBundle=$(kubectl get configmap -n kube-system extension-apiserver-authentication -o=jsonpath='{.data.client-ca-file}' | base64 | tr -d '\n')


kubectl delete secret ${secret} -n ${namespace}
kubectl create secret generic ${secret}  --from-file=tls.crt=${tmpdir}/server-cert.pem --from-file=tls.key=${tmpdir}/server-key.pem --from-literal=caBundle=${caBundle} -n ${namespace}


echo "Create secret captain-webhook-cert in ${namespace}"



# echo $caBundle >> /tmp/k8s-webhook-server/serving-certs/caBundle

 set +e
# Patch the webhook adding the caBundle. It uses an `add` operation to avoid errors in OpenShift because it doesn't set
# a default value of empty string like Kubernetes. Instead, it doesn't create the caBundle key.
# As the webhook is not created yet (the process should be done manually right after this job is created),
# the job will not end until the webhook is patched.
while true; do
  echo "INFO: Trying to patch validating webhook adding the caBundle."
 if kubectl patch validatingwebhookconfiguration "${vwebhook}" --type='json' -p "[{'op': 'add', 'path': '/webhooks/0/clientConfig/caBundle', 'value':'${caBundle}'}]"; then
     break
 fi
 echo "INFO: validating webhook not patched. Retrying in 5s..."
 sleep 5
done



while true; do
  echo "INFO: Trying to patch mutating  webhook adding the caBundle."
 if kubectl patch mutatingwebhookconfiguration "${mwebhook}" --type='json' -p "[{'op': 'add', 'path': '/webhooks/0/clientConfig/caBundle', 'value':'${caBundle}'}]"; then
     break
 fi
 echo "INFO: mutating webhook not patched. Retrying in 5s..."
 sleep 5
done

kubectl annotate secret ${secret} -n ${namespace} cert-created=true
