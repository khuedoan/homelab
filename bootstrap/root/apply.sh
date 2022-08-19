#!/bin/sh

helm template \
    --include-crds \
    --namespace argocd \
    argocd . \
    | kubectl -n argocd apply -f -
