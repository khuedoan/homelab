.POSIX:

env ?= "prod"

default: cluster






cluster:
	ansible-playbook \
		--inventory inventories/${env}.yml \
		cluster.yml

console:
	ansible-console \
		--inventory inventories/${env}.yml
