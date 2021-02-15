.POSIX:

NAME = homelab

default: run ansible

run:
	multipass list | grep $(NAME) \
		|| multipass launch \
		--cpus 4 \
		--mem 8G \
		--disk 64G \
		--cloud-init ./multipass-cloud-init.yaml \
		--name $(NAME) \
		release:20.10
	multipass exec homelab -- sed -i "$$ a\$(shell cat ~/.ssh/id_rsa.pub)" /home/ubuntu/.ssh/authorized_keys

stop:
	multipass stop $(NAME)

clean:
	multipass delete $(NAME)
	multipass purge

ansible:
	python3 -m venv .venv \
		&& . .venv/bin/activate \
		&& pip3 install --upgrade pip \
		&& pip3 install -r requirements.txt \
		&& ansible-playbook --inventory hosts.ini playbook.yml
