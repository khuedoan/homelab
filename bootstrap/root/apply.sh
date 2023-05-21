#!/bin/sh

VALUES="values.yaml"

kubectl get ingress gitea --namespace gitea \
    || VALUES="values-seed.yaml"

helm template \
    --dependency-update \
    --include-crds \
    --namespace argocd \
    --values "${VALUES}" \
    argocd . \
    | kubectl apply -n argocd -f -
