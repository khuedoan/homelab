.POSIX:

default: init apply

init:
	terraform init

plan:
	source ../metal/secrets/env \
	&& terraform plan

apply:
	source ../metal/secrets/env \
	&& terraform apply
