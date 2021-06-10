# TODO

- [(bug) Fix only 15GiB root partition](../metal/roles/pxe-boot/templates/http/kickstart/fedora.ks.j2#L20)
- [(feature) Add lint checks for everything](../Makefile#L29)
- [(feature) Simple script to backup everything](../scripts/backup.sh#L3)
- [(feature) Simple script to restore everything](../scripts/restore.sh#L3)
- [(optimize) Node firewall](../metal/roles/docker/tasks/main.yml#L1)
- [(optimize) Restructure provisioning roles](../metal/roles/docker/tasks/main.yml#L18)
- [(optimize) Use metal values for MetalLB values](../apps/resources/metallb.yaml#L23)
- [(optimize) Use reflector to generate mirrorlist dynamically](../tools/Dockerfile#L3)
