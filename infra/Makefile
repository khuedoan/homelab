.POSIX:
.PHONY: default check graph apply

default: check

check:
	terragrunt hcl validate
	tofu fmt -check -recursive .

graph:
	terragrunt dag graph

apply:
	terragrunt run --all apply
