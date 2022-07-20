#!/bin/sh

VALUES="values.yaml"

curl -fks --connect-timeout 5 https://git.khuedoan.com \
    || VALUES="values-seed.yaml"

helm template \
    --include-crds \
    --namespace argocd \
    --values "${VALUES}" \
    argocd . \
    | kubectl apply -n argocd -f -
