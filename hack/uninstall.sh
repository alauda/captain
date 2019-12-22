	kubectl delete mutatingwebhookconfigurations captain-mutating-webhook-configuration
	kubectl delete validatingwebhookconfigurations captain-validating-webhook-configuration
	kubectl delete crd helmrequests.app.alauda.io releases.app.alauda.io chartrepos.app.alauda.io charts.app.alauda.io
	kubectl delete ns captain-system
