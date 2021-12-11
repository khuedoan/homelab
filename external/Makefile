.POSIX:

default: apply

~/.terraform.d/credentials.tfrc.json:
	terraform login

.terraform.lock.hcl: ~/.terraform.d/credentials.tfrc.json versions.tf
	terraform init

apply: .terraform.lock.hcl
	terraform apply
