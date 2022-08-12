#!/bin/sh

echo "WARNING: ArgoCD admin can do anything in the cluster, only use it for just enough initial setup or in emergencies." >&2
export KUBECONFIG=./metal/kubeconfig.yaml
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d
