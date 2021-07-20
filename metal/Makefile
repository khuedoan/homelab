.POSIX:

default: run

run:
	ansible-playbook \
		--inventory hosts.yml \
		main.yml

lint:
	ansible-lint --project-dir .
