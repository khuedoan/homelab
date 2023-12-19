.POSIX:

default: apply

terraform.tfvars:
	ansible-playbook tfvars.yml

.terraform.lock.hcl: versions.tf terraform.tfvars
	terraform init
	touch .terraform.lock.hcl

namespaces:
	ansible-playbook namespaces.yml

plan: .terraform.lock.hcl
	terraform plan

apply: .terraform.lock.hcl namespaces
	terraform apply -auto-approve
