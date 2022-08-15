#!/bin/sh

helm template \
    --include-crds \
    --namespace argocd \
    argocd . \
    | kubectl apply -f -
