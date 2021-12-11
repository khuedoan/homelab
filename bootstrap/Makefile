.POSIX:
.EXPORT_ALL_VARIABLES:

KUBECONFIG = ../metal/kubeconfig.yaml

default: argocd root-app

.PHONY: argocd
argocd:
	kustomize build ./argocd | kubectl apply -f -
	# TODO wait for argocd

.PHONY: root-app
root-app:
	kustomize build ./root-app | kubectl apply -f -
