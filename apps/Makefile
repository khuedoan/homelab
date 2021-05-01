.POSIX:

default: apply

apply:
	kubectl --kubeconfig ../infra/kube_config.yaml --namespace argocd apply --filename .

delete:
	kubectl --kubeconfig ../infra/kube_config.yaml --namespace argocd delete --filename .
