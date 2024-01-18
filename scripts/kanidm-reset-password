#!/bin/sh

set -eu

account="${1}"

echo "WARNING: Kanidm admin can do anything in the cluster, only use it for just enough initial setup or in emergencies." >&2
export KUBECONFIG=./metal/kubeconfig.yaml
kubectl exec -it -n kanidm statefulset/kanidm -- kanidmd recover-account "${account}"
