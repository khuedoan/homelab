.POSIX:

default: init metal infra

init:
	vagrant up

metal:
	python3 -m venv .venv \
		&& . .venv/bin/activate \
		&& pip3 install --upgrade pip \
		&& pip3 install -r requirements.txt \
		&& ansible-playbook --inventory hosts.ini playbook.yml

infra:
	echo infra

stop:
	vagrant halt

clean:
	vagrant destroy -f
