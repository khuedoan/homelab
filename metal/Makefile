.POSIX:

default: init run

init:
	python3 -m venv .venv \
		&& . .venv/bin/activate \
		&& pip3 install --upgrade pip \
		&& pip3 install -r requirements.txt

run:
	. .venv/bin/activate \
		&& ansible-playbook --ask-become-pass --inventory hosts.ini playbook.yml
