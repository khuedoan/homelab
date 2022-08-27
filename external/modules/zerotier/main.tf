locals {
  router_ip = cidrhost(var.managed_route, 1) # Use the second IP in the VPN subnet as the router
}

resource "zerotier_network" "network" {
  name        = var.name
  description = var.description
  private     = true

  route {
    target = var.managed_route
  }

  dynamic "route" {
    for_each = var.bridged_routes

    content {
      target = route.value
      via    = local.router_ip
    }
  }

  assignment_pool {
    start = cidrhost(var.managed_route, 0)
    end   = cidrhost(var.managed_route, -1)
  }
}

resource "zerotier_identity" "router" {}

resource "zerotier_member" "router" {
  network_id              = zerotier_network.network.id
  name                    = "router"
  member_id               = zerotier_identity.router.id
  allow_ethernet_bridging = true
  no_auto_assign_ips      = true
  ip_assignments = [
    local.router_ip
  ]
}

resource "kubernetes_secret" "router" {
  metadata {
    name      = "zerotier-router"
    namespace = "zerotier"
  }

  data = {
    ZEROTIER_NETWORK_ID      = zerotier_network.network.id
    ZEROTIER_IDENTITY_PUBLIC = zerotier_identity.router.public_key
    ZEROTIER_IDENTITY_SECRET = zerotier_identity.router.private_key
  }
}
