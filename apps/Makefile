.POSIX:
.EXPORT_ALL_VARIABLES:

KUBECONFIG = ../infra/kube_config.yaml

default: apply

apply:
	kubectl apply --kustomize .

delete:
	kubectl delete --kustomize .
