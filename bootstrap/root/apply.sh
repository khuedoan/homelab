#!/bin/sh

helm template \
    --include-crds \
    --namespace argocd \
    argocd . \
    | kubectl apply -n argocd -f -
