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

rm -rf /tmp/captain-test

git clone https://github.com/alauda/captain-test-charts /tmp/captain-test

cd /tmp/captain-test || exit

## Test chartrepo
yellow "[0] TEST CHARTREPO"
# for mac add a .bak...
sed  -i .bak 's/alauda-system/captain/g' chartrepo.yaml
kubectl apply -f chartrepo.yaml -n captain

until [ $(kubectl get ctr -n captain |grep captain-test  | awk '{print $3}') == "Synced" ]
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


kubectl delete -f chartrepo.yaml
kubectl delete -f hr/basic/
kubectl delete -f hr/dep/
rm -rf /tmp/captain-test