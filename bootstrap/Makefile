.POSIX:
.EXPORT_ALL_VARIABLES:

KUBECONFIG = ../metal/kubeconfig.yaml

default: argocd root

.PHONY: argocd
argocd:
	kustomize build ./argocd | kubectl apply -f -
	# TODO wait for argocd

.PHONY: root
root:
	kustomize build ./root | kubectl apply -f -
