provider "lxd" {
}

resource "lxd_container" "test1" {
  name      = "test1"
  remote    = "images"
  image     = "ubuntu/18.04"
  ephemeral = false

  config = {
    "boot.autostart" = true
  }

  limits = {
    cpu = 2
  }
}
