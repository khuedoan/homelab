.POSIX:

default: run

run:
	ansible-playbook \
		--inventory hosts.yml \
		main.yml

shutdown:
	ansible-playbook \
		--inventory hosts.yml \
		shutdown.yml

lint:
	ansible-lint --project-dir .
