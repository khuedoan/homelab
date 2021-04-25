.POSIX:

default: apply

apply:
	kubectl --kubeconfig ../infra/kube_config.yaml apply -f .

delete:
	kubectl --kubeconfig ../infra/kube_config.yaml delete -f .
