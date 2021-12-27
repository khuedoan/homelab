.POSIX:

env ?= "prod"

default: boot cluster

~/.ssh/id_ed25519:
	ssh-keygen -t ed25519 -f "$@"

boot: ~/.ssh/id_ed25519
	ansible-playbook \
		--inventory inventories/${env}.ini \
		boot.yml

cluster:
	ansible-playbook \
		--inventory inventories/${env}.ini \
		cluster.yml

shutdown:
	ansible-playbook \
		--inventory inventories/${env}.ini \
		shutdown.yml

vagrant:
	vagrant up
