.POSIX:

default: namespace argocd root

namespace:
	kubectl create namespace argocd --dry-run=client --output=yaml \
		| kubectl apply -f -

.PHONY: argocd
argocd:
	for sm in $$(kubectl get servicemonitors -n argocd -o name); do \
		kubectl annotate $$sm -n argocd meta.helm.sh/release-name=argocd --overwrite; \
		kubectl annotate $$sm -n argocd meta.helm.sh/release-namespace=argocd --overwrite; \
	done
	helm upgrade --install argocd argocd \
	--namespace argocd \
	--dependency-update
	cd argocd && ./apply.sh

.PHONY: root
root:
	cd root && ./apply.sh
