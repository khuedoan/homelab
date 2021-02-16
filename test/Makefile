.POSIX:

NAME = metal

default: init metal infra

init:
	for count in 0 1; do \
		multipass list | grep $(NAME)$$count \
		|| multipass launch \
			--cpus 2 \
			--mem 4G \
			--disk 64G \
			--cloud-init ./multipass-cloud-init.yaml \
			--name $(NAME)$$count \
			release:20.10; \
		multipass exec $(NAME)$$count -- sed -i "$$ a\$(shell cat ~/.ssh/id_rsa.pub)" /home/ubuntu/.ssh/authorized_keys; \
	done

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
	for count in 0 1; do \
		multipass delete $(NAME)$$count; \
	done
	multipass purge
