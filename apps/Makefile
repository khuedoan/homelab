.POSIX:

default: init apply

init:
	kubectl create namespace argocd --dry-run=client -o yaml | kubectl --kubeconfig ../infra/kube_config.yaml apply -f -
	kubectl --kubeconfig ../infra/kube_config.yaml apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml

apply:
	kubectl --kubeconfig ../infra/kube_config.yaml --namespace argocd apply --filename .

delete:
	kubectl --kubeconfig ../infra/kube_config.yaml --namespace argocd delete --filename .
