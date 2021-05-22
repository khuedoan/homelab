.POSIX:
.EXPORT_ALL_VARIABLES:

KUBECONFIG = ../infra/kube_config.yaml

default: init apply

init:
	kubectl create namespace argocd --dry-run=client -o yaml | kubectl apply -f -
	kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml

apply:
	kubectl --namespace argocd apply --filename .

delete:
	kubectl --namespace argocd delete --filename .
