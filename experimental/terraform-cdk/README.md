# Terraform CDK

## Usage

Compile:

```sh
go build              # Builds your go project
```

Synthesize:

```sh
cdktf synth [stack]   # Synthesize Terraform resources to cdktf.out/
```

Diff:

```sh
cdktf diff [stack]    # Perform a diff (terraform plan) for the given stack
```

Deploy:

```sh
cdktf deploy [stack]  # Deploy the given stack
```

Destroy:

```sh
cdktf destroy [stack] # Destroy the given stack
```
