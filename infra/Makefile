.POSIX:
.EXPORT_ALL_VARIABLES:

TF_IN_AUTOMATION = true

default: init apply

init:
	terraform init -input=false -backend-config=backend.tfvars

plan:
	terraform plan -input=false -out=/tmp/tfplan

apply: plan
	terraform apply -input=false /tmp/tfplan

destroy:
	terraform destroy
