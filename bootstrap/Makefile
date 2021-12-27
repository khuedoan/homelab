.POSIX:
.EXPORT_ALL_VARIABLES:

KUBECONFIG = ../metal/kubeconfig.yaml
NAMESPACE = argocd

default: namespace apply

charts: Chart.yaml
	helm dependency update

namespace:
	kubectl create namespace argocd --dry-run=client --output=yaml \
		| kubectl apply -f -

apply: charts
	# ArgoCD only uses Helm as YAML generator
	helm template \
		--include-crds \
		--namespace ${NAMESPACE} \
		argocd . \
		| kubectl apply -n argocd -f -
