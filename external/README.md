# External resources

This layer will deploy resources that require external dependencies using the following provisioners:

Terraform:

- Create external resources
- Add external secrets to namespaces
- Create an ApplicationSet

ArgoCD (via the ApplicationSet created by Terraform):

- Deploy Helm charts in the subdirectories
