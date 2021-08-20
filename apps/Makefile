.POSIX:
.EXPORT_ALL_VARIABLES:

KUBECONFIG = ../cluster/kubeconfig.yaml

default: apply

apply:
	kubectl apply --kustomize .

destroy:
	kubectl delete --kustomize .
