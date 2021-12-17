.POSIX:
.EXPORT_ALL_VARIABLES:

KUBECONFIG = ../metal/kubeconfig.yaml

default: argocd root

.PHONY: argocd
argocd:
	kustomize build ./argocd | kubectl apply -f -
	kubectl -n argocd wait --for condition=established \
		crd/applications.argoproj.io \
		crd/applicationsets.argoproj.io

.PHONY: root
root:
	kustomize build ./root | kubectl apply -f -
