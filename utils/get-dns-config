#!/bin/sh

export KUBECONFIG=./metal/kubeconfig.yaml
kubectl get ingress --all-namespaces --no-headers --output custom-columns="ADDRESS:.status.loadBalancer.ingress[0].ip,HOST:.spec.rules[0].host"
