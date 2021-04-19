# TODO

- [check if this is a bug](../infra/platform/cluster.tf#L85)
- [use Ansible wait_for /var/run/docker.sock](../infra/platform/cluster.tf#L105)
- [remote backend (etcd or minio)](../infra/platform/terraform.tf#L2)
- [convert to YAML for Terraform yamldecode](../metal/hosts.ini#L1)
- [Optimize SELinux](../metal/roles/lxd/tasks/main.yml#L1)
- [Optimize firewall](../metal/roles/lxd/tasks/main.yml#L6)
- [change to /var/lib/lxd/server.crt after https](../metal/roles/lxd/tasks/main.yml#L26)
- [use btrfs in k8s 1.19.8 https](../metal/roles/lxd/templates/leader.yaml.j2#L17)
