# Annotations Captain Use

Captain use annotations to do some extra works, this document will introduce each of them


## `captain-no-sync`
Works on: `HelmRequest`

Values: True/False

Description:
	If you do not want captain to sync a HelmRequest anymore, and you do not want to delete it for now, you can use this annotation to tell captain this HelmRequest skip processing this HelmRequest.
	For example, if you have accidentally create two HelmRequest which use the same chart and target to the same namespace, and the resources are already installed in this namespace. In this situation, delete either of this HelmRequest will delete the contained resource. This annotation can help you with this.(Of course there are ways to safely delete a HelmRequest without delete the contained resource)

## `kubectl-captain.resync`
Works on: `HelmRequest`

Values: timestamp

Description:
	[kubectl-captain](https://github.com/alauda/kubectl-captain) has a sub-command called `trigger-update`, it will force captain to resync a HelmRequest, this is very convenient for many user cases. To do this ,it will add this annotations to the taget HelmRequst. Uses can also add arbitrarily annotations to trigger resync on a HelmRequest

## `cpaas.io/last-sync-at`
Works on: `ChartRepo`

Values: timestamp

Description:
	This annotations indicate when is the last time captain sync this ChartRepo. Usually chartrepos will receive charts update now and then, captain periodically poll updates from it's `index.yaml`. This timestamp can show users whether captain is working normally doing this.
