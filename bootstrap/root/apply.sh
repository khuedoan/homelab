#!/bin/sh

set -eu

ADDITIONAL_VALUES=""

# If the Gitea service does not exist, it means that this is the first installation, clone from a seed repo instead
# ArgoCD will automatically switch to Gitea once the repository is clonable
if ! kubectl get service gitea-http --namespace gitea; then
    # Use Ansible to get a consistent output with ../metal
    CONTROLLER_IP="$(\
        ANSIBLE_LOAD_CALLBACK_PLUGINS=true \
        ANSIBLE_STDOUT_CALLBACK=json \
        ansible localhost -m ansible.builtin.setup -a 'filter=ansible_default_ipv4' \
        | jq --raw-output '.plays[0].tasks[0].hosts.localhost.ansible_facts.ansible_default_ipv4.address' \
    )"

    # Create a temporary git server
    SEED_REPO="${PWD}/seed-repo/homelab.git"
    [ -d "${SEED_REPO}" ] || git clone --bare ../.. "${SEED_REPO}"
    git -C "${SEED_REPO}" update-server-info
    git -C "${SEED_REPO}" fetch --all
	docker start homelab-seed-repo || docker run \
		--volume "${PWD}/seed-repo:/usr/share/nginx/html:ro" \
		--publish 8000:80 \
		--rm \
		--detach \
		--name homelab-seed-repo \
		nginx:latest

    ADDITIONAL_VALUES="gitops.repo=http://${CONTROLLER_IP}:8000/homelab.git"
fi

echo "${ADDITIONAL_VALUES}"

helm template \
    --include-crds \
    --namespace argocd \
    --set "${ADDITIONAL_VALUES}" \
    argocd . \
    | kubectl apply -n argocd -f -
