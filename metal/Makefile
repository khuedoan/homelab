.POSIX:

env ?= "prod"

default: boot cluster

boot:
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

lint:
	ansible-lint --project-dir .

vagrant:
	vagrant up
