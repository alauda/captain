# Captain使用注解`Annotations`说明

`Captain`使用注解做了一些其它的功能实现，本文档将一一进行介绍。

## `captain-no-sync`
适用于: `HelmRequest`

可取值: True/False

描述:
    如果你不想让`Captain`再同步一个`HelmRequest`，并且你暂时不想删除它，你可以使用这个注解告诉`Captain` 跳过处理这个`HelmRequest`。
    例如，如果您不小心创建了两个`HelmRequest`，它们使用相同的`Chart`、`namespace`，并且资源已经部署在这个命名空间中。
    在这种情况下，删除其中的任意一个`HelmRequest`都会删除相应的资源。
    这个注解可以帮助你解决这个问题。（当然有一些方法可以安全地删除`HelmRequest`而不删除包含的资源）

## `captain-keep-resources`
适用于: `HelmRequest`

可取值: True/False

描述:
	如果你想在通过helm卸载chart时保留已部署的k8s资源，你可以使用这个注释告诉`Captain` ，这个HelmRequest对应的chart在进行卸载时将会保留资源。

## `captain-force-adopt-resources`
适用于: `HelmRequest`

可取值: True/False

描述:
	如果你想在通过helm安装或升级chart版本时领养k8s资源，你可以使用这个注释告诉`Captain`，这个HelmRequest将在安装或升级对应的chart时强制领养资源。因为在最新的helm版本中，不允许更新不属于当前release的同名资源。

## `kubectl-captain.resync`
适用于: `HelmRequest`

可取值: timestamp

描述:
	[kubectl-captain](https://github.com/alauda/kubectl-captain) 有一个子命令叫做`trigger-update`，它会强制`Captain`重新同步一个`HelmRequest`，添加该注释至`HelmRequst`中可以实现重新同步的效果，这对于很多用户来说非常方便。
	用户还可以任意添加注解来触发`HelmRequest`上的重新同步。

## `cpaas.io/last-sync-at`
适用于: `ChartRepo`

可取值: timestamp

描述:
	该注释标识`Captain`最后一次同步`ChartRepo`的时间。通常`chartrepos`会不时收到`Chart`更新，`Captain`定期从它的`index.yaml`中轮询更新。
	这个时间戳可以向用户显示`Captain`是否正常工作。
