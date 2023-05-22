.POSIX:

env ?= "prod"

default: k3d

~/.ssh/id_ed25519:
	ssh-keygen -t ed25519 -P '' -f "$@"

boot: ~/.ssh/id_ed25519
	ansible-playbook \
		--inventory inventories/${env}.yml \
		boot.yml

cluster:
	ansible-playbook \
		--inventory inventories/${env}.yml \
		cluster.yml

console:
	ansible-console \
		--inventory inventories/${env}.yml

k3d:
	k3d cluster start homelab-dev || k3d cluster create --config k3d-dev.yaml
	k3d kubeconfig get homelab-dev > kubeconfig.yaml
