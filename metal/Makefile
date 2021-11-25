.POSIX:

default: boot cluster

boot:
	ansible-playbook \
		--inventory hosts.yml \
		main.yml

cluster:
	ansible-playbook \
		--inventory hosts.yml \
		cluster.yml

shutdown:
	ansible-playbook \
		--inventory hosts.yml \
		shutdown.yml

lint:
	ansible-lint --project-dir .
