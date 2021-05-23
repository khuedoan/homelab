.POSIX:
.EXPORT_ALL_VARIABLES:

KUBECONFIG = ../infra/kube_config.yaml

default: apply

apply:
	kubectl apply --kustomize .

destroy:
	kubectl delete --kustomize .
