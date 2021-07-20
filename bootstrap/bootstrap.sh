#!/bin/sh

set -e

# Create ephemeral cluster
kind create cluster \
    --wait 60s \
    --config kind.yaml

# Install Sidero
clusterctl init \
    --bootstrap talos \
    --control-plane talos \
    --infrastructure sidero \
    --config clusterctl.yaml
kubectl wait pods \
    --all \
    --all-namespaces \
    --for=condition=Ready

# Apply server classes
kubectl apply \
    --filename serverclasses/

# Waker servers up
wol '00:23:24:d1:f3:f0'
# wol '00:23:24:d1:f4:d6'
# wol '00:23:24:d1:f5:69'
# wol '00:23:24:e7:04:60'

# Create cluster
export CONTROL_PLANE_ENDPOINT=$(kubectl get server -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}')
clusterctl config cluster homelab \
    --infrastructure sidero \
    --config clusterctl.yaml > cluster/homelab.yaml
kubectl apply \
    --filename cluster/homelab.yaml

# Get kube config
clusterctl get kubeconfig homelab > kubeconfig.yaml
