#!/bin/sh

curl -fks --connect-timeout 5 https://git.khuedoan.com \
    || extra_args="--values values-seed.yaml"

helm template \
    --include-crds \
    --namespace argocd \
    ${extra_args} \
    argocd . \
    | kubectl apply -n argocd -f -
