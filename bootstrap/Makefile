.POSIX:
.EXPORT_ALL_VARIABLES:

KUBECONFIG = ../cluster/kubeconfig.yaml

default: argocd root-app

.PHONY: argocd
argocd:
	helm dependencies update ./argocd
	helm template ./argocd | kubectl apply -f -
	# TODO wait for argocd

.PHONY: root-app
root-app:
	kustomize build ./root-app | kubectl apply -f -
