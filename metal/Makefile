.POSIX:

default: boot cluster

boot:
	ansible-playbook \
		--inventory inventories/prod.yml \
		main.yml

cluster:
	ansible-playbook \
		--inventory inventories/prod.yml \
		cluster.yml

shutdown:
	ansible-playbook \
		--inventory inventories/prod.yml \
		shutdown.yml

lint:
	ansible-lint --project-dir .

dev:
	vagrant up
	ansible-playbook \
		--inventory inventories/dev.yml \
		cluster.yml
