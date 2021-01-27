#TODO
.POSIX:

default: help

help:
	@echo "Khue's homelab"

platform:
	cd ./platform \
		&& terraform init \
		&& terraform apply

apps:
	echo hello

reset-infra:
	./infra/scripts/reset.py
