.POSIX:

default: namespace argocd root

argocd/charts: argocd/Chart.yaml
	cd argocd \
		&& helm dependency update

namespace:
	kubectl create namespace argocd --dry-run=client --output=yaml \
		| kubectl apply -f -

.PHONY: argocd
argocd: argocd/charts
	cd argocd && \
		./apply.sh

.PHONY: root
root:
	cd root && \
		./apply.sh
