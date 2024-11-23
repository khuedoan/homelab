.POSIX:

default: apply

~/.terraform.d/credentials.tfrc.json:
	tofu login

terraform.tfvars:
	cp terraform.tfvars.example ${@}
	nvim ${@}

.terraform.lock.hcl: ~/.terraform.d/credentials.tfrc.json versions.tf terraform.tfvars
	tofu init
	touch ${@}

namespaces:
	ansible-playbook namespaces.yml

plan: .terraform.lock.hcl
	tofu plan

apply: .terraform.lock.hcl namespaces
	tofu apply -auto-approve
