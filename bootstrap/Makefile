.POSIX:
.EXPORT_ALL_VARIABLES:

KUBECONFIG = ../cluster/kubeconfig.yaml

default: argocd root-app

argocd:
	kustomize build ./system/argocd | kubectl apply -f -
	# TODO wait for argocd

root-app:
	kustomize build ./system/root-app | kubectl apply -f -
