.POSIX:
.EXPORT_ALL_VARIABLES:

KUBECONFIG = ../metal/kubeconfig.yaml

default: argocd root

.PHONY: argocd
argocd:
	kustomize build ./argocd | kubectl apply -f -

.PHONY: root
root:
	kustomize build ./root | kubectl apply -f -
