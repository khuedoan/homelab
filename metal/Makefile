.POSIX:

default: boot cluster

boot:
	ansible-playbook \
		--inventory inventories/prod.ini \
		main.yml

cluster:
	ansible-playbook \
		--inventory inventories/prod.ini \
		cluster.yml

shutdown:
	ansible-playbook \
		--inventory inventories/prod.ini \
		shutdown.yml

lint:
	ansible-lint --project-dir .

dev:
	vagrant up
	ansible-playbook \
		--inventory inventories/dev.ini \
		cluster.yml
