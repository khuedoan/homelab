.POSIX:

default: init apply

init:
	terraform init

plan:
	terraform plan

apply:
	# TODO (bug) Apply LXD VMs in parallel
	# Currently have to turn it off because it will not spread evenly accross LXD nodes
	terraform apply --parallelism 1
