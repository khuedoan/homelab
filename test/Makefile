.POSIX:

NAME = homelab

run:
	multipass list | grep $(NAME) \
	|| multipass launch \
		--cpus 4 \
		--mem 8G \
		--disk 64G \
		--cloud-init ./multipass-cloud-init.yaml \
		--name $(NAME) \
		release:20.10
	multipass shell $(NAME)

stop:
	multipass stop $(NAME)

clean:
	multipass delete $(NAME)
	multipass purge
