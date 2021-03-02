.POSIX:

default: init provision

init:
	python3 -m venv .venv \
		&& . .venv/bin/activate \
		&& pip3 install --upgrade pip \
		&& pip3 install -r requirements.txt

reset:
	. .venv/bin/activate \
		&& ansible-playbook --ask-become-pass --inventory hosts.ini playbook.yml

provision:
	. .venv/bin/activate \
		&& ansible-playbook --inventory hosts.ini playbook.yml --tags provision
