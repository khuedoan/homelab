.POSIX:

run:
	multipass list | grep homelab \
	|| multipass launch \
		--cpus 4 \
		--mem 8G \
		--disk 64G \
		--cloud-init ./multipass-cloud-init.yaml \
		--name homelab \
		release:20.10
	multipass shell homelab

clean:
	multipass delete homelab
	multipass purge
