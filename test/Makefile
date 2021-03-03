.POSIX:

default: init snapshot metal infra

init:
	vagrant up

snapshot:
	vagrant snapshot save init

metal:
	python3 -m venv .venv \
		&& . .venv/bin/activate \
		&& pip3 install --upgrade pip \
		&& pip3 install -r requirements.txt \
		&& ansible-playbook --inventory hosts.ini main.yml

infra:
	echo infra

stop:
	vagrant halt

restore:
	vagrant snapshot restore init

clean:
	vagrant destroy -f
