# Development Container for Homelab

This directory contains configuration files for a Visual Studio Code Development Container. The development container provides a consistent, isolated environment with all the necessary tools pre-installed to work on the homelab project.

## Prerequisites

- [Visual Studio Code](https://code.visualstudio.com/)
- [Docker](https://www.docker.com/products/docker-desktop)
- [Remote - Containers extension](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers) for VS Code

## Getting Started

1. Clone the repository
2. Place your SSH keys in the `.ssh` directory at the root of the repository
   - These will be mounted to `/root/.ssh` inside the container
   - This is necessary for tasks like SSH access to your homelab nodes
3. Open the repository in VS Code
4. Click the green button in the bottom-left corner and select "Reopen in Container"
5. Wait for the container to build and start (this may take a few minutes the first time)

## Included Tools

- Docker (in Docker)
- Ansible and related tools
- kubectl, Helm, and k9s for Kubernetes management
- pre-commit for managing pre-commit hooks
- Nix and direnv for development environments
- Make, Python, and other common development tools

## Notes

- The container uses host networking, so services you run inside the container will be accessible at `localhost`
- The container runs with elevated privileges (needed for Docker-in-Docker)
- Your SSH keys will be available inside the container, allowing you to connect to your homelab servers
- KUBECONFIG is automatically set to point to `./metal/kubeconfig.yaml`

## VS Code Extensions

The development container comes with several useful VS Code extensions pre-installed:

- Terraform extension for working with Terraform files
- Kubernetes tools for working with Kubernetes resources
- YAML support with Kubernetes schema validation
- Makefile tools for better Makefile support
- Markdown support for documentation
- Nix language support
- Bash IDE for shell script development
