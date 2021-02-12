# Layer 0

Bare metal provisioning

- Render PXE boot configs
- Start Docker based PXE server
- Turn off all nodes
- Wake them up using magic packet
- Install OS using PXE and kickstart
- Wait for the OS installation to finish and reboot to the new OS
- Start basic provisioning
  - Install Docker
  - Create simple etcd container for Terraform state
- Generate Terraform backend config
