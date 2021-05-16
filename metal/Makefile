.POSIX:

default: run

run:
	ansible-playbook \
		--inventory hosts.ini \
		main.yml

shutdown:
	ansible-playbook \
		--inventory hosts.ini \
		shutdown.yml
