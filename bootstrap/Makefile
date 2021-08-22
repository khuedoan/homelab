.POSIX:
.EXPORT_ALL_VARIABLES:

KUBECONFIG = ../cluster/kubeconfig.yaml

default: argocd root-app

.PHONY: argocd
argocd:
	kustomize build ./system/argocd | kubectl apply -f -
	# TODO wait for argocd

.PHONY: root-app
root-app:
	kustomize build ./system/root-app | kubectl apply -f -
