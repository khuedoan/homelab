.POSIX:
.EXPORT_ALL_VARIABLES:

TF_IN_AUTOMATION = true

default: init plan apply

init:
	terraform init -input=false

plan:
	terraform plan -input=false -out=/tmp/tfplan

apply:
	terraform apply -input=false /tmp/tfplan

destroy:
	terraform destroy
