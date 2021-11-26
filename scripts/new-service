#!/bin/sh

mkdir -p "apps/${1}"

cat << EOT > "apps/${1}/Chart.yaml"
apiVersion: v2
name: CHANGEME
version: 0.0.0
dependencies:
- name: CHANGEME
  version: CHANGEME
  repository: CHANGEME
EOT

touch "apps/${1}/values.yaml"
