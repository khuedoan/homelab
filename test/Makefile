.POSIX:

NAME = homelab

default: init metal infra

init:
	multipass list | grep $(NAME) \
		|| multipass launch \
		--cpus 4 \
		--mem 8G \
		--disk 64G \
		--cloud-init ./multipass-cloud-init.yaml \
		--name $(NAME) \
		release:20.10
	multipass exec homelab -- sed -i "$$ a\$(shell cat ~/.ssh/id_rsa.pub)" /home/ubuntu/.ssh/authorized_keys

shell:
	multipass shell $(NAME)

metal:
	python3 -m venv .venv \
		&& . .venv/bin/activate \
		&& pip3 install --upgrade pip \
		&& pip3 install -r requirements.txt \
		&& ansible-playbook --inventory hosts.ini playbook.yml

infra:
	echo infra

stop:
	multipass stop $(NAME)

clean:
	multipass delete $(NAME)
	multipass purge
