.POSIX:
.EXPORT_ALL_VARIABLES:

KUBECONFIG = ../metal/kubeconfig.yaml

default: argocd

.PHONY: argocd
argocd:
	kustomize build ./argocd | kubectl apply -f -
