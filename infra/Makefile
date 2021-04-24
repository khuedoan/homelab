.POSIX:

default: init apply

init:
	cd platform \
	&& terraform init

plan:
	platform \
	&& terraform plan

apply:
	cd platform \
	&& terraform apply
