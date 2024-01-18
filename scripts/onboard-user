#!/bin/sh

username="${1}"
fullname="${2}"
mail="${3}"

export KUBECONFIG=./metal/kubeconfig.yaml
host="$(kubectl get ingress --namespace kanidm kanidm --output jsonpath='{.spec.rules[0].host}')"

kanidm person create "${username}" "${fullname}" --url "https://${host}" --name idm_admin
kanidm person update "${username}" --url "https://${host}" --name idm_admin --mail "${mail}"
# TODO better group management
kanidm group add-members "editor" "${username}" --url "https://${host}" --name idm_admin
kanidm person credential create-reset-token "${username}" --url "https://${host}" --name idm_admin
