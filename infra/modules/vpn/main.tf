resource "lxd_container" "vpn" {
  name      = "vpn"
  image     = "ubuntu:20.04"
  ephemeral = false
}
