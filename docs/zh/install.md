# 安装指南


## 安装

```bash
kubectl create ns captain-system
kubectl create clusterrolebinding captain --serviceaccount=captain-system:default --clusterrole=cluster-admin
kubectl apply -n captain-system -f https://raw.githubusercontent.com/alauda/captain/master/artifacts/all/deploy.yaml
```

### 使用主分支构建镜像
captain使用github的持续集成服务action为主分支自动生成镜像。如果您想使用最新的镜像，请更新yaml文件，并替换captain镜像为

```bash
alaudapublic/captain:latest
```




## 卸载
```bash
kubectl delete -n  captain-system -f https://raw.githubusercontent.com/alauda/captain/master/artifacts/all/deploy.yaml
kubectl delete clusterrolebinding captain
kubectl delete ns captain-system
```
