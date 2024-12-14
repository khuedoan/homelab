.POSIX:

export KUBECONFIG = $(shell pwd)/../metal/kubeconfig-${env}.yaml

.PHONY: bootstrap
bootstrap:
	ansible-playbook \
		--extra-vars "env=${env}" \
		bootstrap.yml
