package main

import (
	"github.com/aws/constructs-go/constructs/v3"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
)

func NewMyStack(scope constructs.Construct, id string) cdktf.TerraformStack {
	stack := cdktf.NewTerraformStack(scope, &id)

	// The code that defines your stack goes here

	return stack
}

func main() {
	app := cdktf.NewApp(nil)

	NewMyStack(app, "cdk")

	app.Synth()
}
