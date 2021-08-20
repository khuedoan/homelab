.POSIX:
.EXPORT_ALL_VARIABLES:

KUBECONFIG = ../infra/kubeconfig.yaml

default: apply

apply:
	kubectl apply --kustomize .

destroy:
	kubectl delete --kustomize .
