.POSIX:

default: run

run:
	ansible-playbook \
		--ask-become-pass \
		--inventory hosts.ini \
		main.yml

shutdown:
	ansible-playbook \
		--inventory hosts.ini \
		shutdown.yml
