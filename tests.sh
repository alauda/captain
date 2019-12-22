#!/usr/bin/env bash
#/usr/bin/env bash

NORMAL=$(tput sgr0)
GREEN=$(
  tput setaf 2
  tput bold
)
YELLOW=$(tput setaf 3)
RED=$(tput setaf 1)

function red() {
  echo -e "$RED$*$NORMAL"
}

function green() {
  echo -e "$GREEN$*$NORMAL"
}

function yellow() {
  echo -e "$YELLOW$*$NORMAL"
}


kubectl create ns captain-system


rm -rf /tmp/captain-test

git clone https://github.com/alauda/captain-test-charts /tmp/captain-test

cd /tmp/captain-test || exit

## Test chartrepo
yellow "[0] TEST CHARTREPO"
# for mac add a .bak...
sed  -i .bak 's/alauda-system/captain-system/g' chartrepo.yaml
kubectl apply -f chartrepo.yaml -n captain-system

until [ $(kubectl get ctr -n captain-system | grep captain-test  | awk '{print $3}') == "Synced" ]
do
    green "Wating chartrepo to be ready...."
    sleep 1
done

yellow "ChartRepo captain-test synced"

## Test false dep
yellow "[1] TEST FALSE DEP"
kubectl apply -f hr/dep/dep.yaml

until  kubectl describe hr dep-jenkins  |grep  "FailedSync" |grep "not found"
do
  green "Wating for dep check error ..."
  sleep 1
done


## Test basic
yellow "[2] TEST BASIC"
kubectl apply -f hr/basic/basic.yaml

until [ $(kubectl get hr  basic-nginx-ingress -o json | jq -r .status.phase) == "Synced" ]
do
  green "Wating for hr/basic-nginx-ingress synced..."
  sleep 1
done

yellow "HelmRequest basic-nginx-ingress synced"


## Test crd-install
yellow "[3] TEST CRD-INSTALL"
kubectl apply -f hr/crd-cr/cr.yaml

until [ $(kubectl get hr  tomcat-crd-install -o json | jq -r .status.phase) == "Synced" ]
do
  green "Wating for hr/tomcat-crd-install synced..."
  sleep 1
done

yellow "HelmRequest tomcat-crd-install synced..."

# delete the crd
kubectl delete -f hr/crd-cr/cr.yaml
# kubectl delete crd 
kubectl delete crd crontabs.stable.example.com

## Test cr install (without CRD)
yellow "[4] TEST ONLY CR INSTALL"
kubectl apply -f hr/bad-cr/bad-cr.yaml

until [ $(kubectl get hr  ghost-bad-cr -o json | jq -r .status.phase) == "Failed" ]
do
  green "Wating for hr/ghost-bad-cr faild..."
  sleep 1
done

yellow "Helmrequest ghost-bad-cr failed..."





kubectl delete ctr  captain-test -n captain-system
kubectl delete -f hr/basic/
kubectl delete -f hr/dep/
kubectl delete -f hr/bad-cr
rm -rf /tmp/captain-test
