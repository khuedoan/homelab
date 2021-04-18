.POSIX:

default: init apply

init:
	cd platform \
	&& terraform init

plan:
	source ../metal/secrets/env \
	&& cd platform \
	&& terraform plan

apply:
	source ../metal/secrets/env \
	&& cd platform \
	&& terraform apply
