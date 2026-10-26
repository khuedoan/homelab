resource "kubectl_manifest" "wireguard" {
  depends_on        = [kubectl_manifest.openbao]
  server_side_apply = true
  yaml_body = yamlencode({
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = "wireguard"
      namespace = "argocd"
      labels    = { "app.kubernetes.io/part-of" = "homelab" }
    }
    spec = {
      project = "default"
      sources = [{
        repoURL        = "ghcr.io/h44z/charts"
        targetRevision = "0.7.3"
        chart          = "wg-portal"
        helm = {
          releaseName = "wireguard"
          values = yamlencode({
            fullnameOverride = "wireguard"
            image            = { tag = "v2.3.1" }
            strategy         = { type = "Recreate" }
            resources        = { requests = { cpu = "50m", memory = "64Mi" } }
            persistence      = { enabled = true }
            envFrom          = [{ secretRef = { name = "wireguard-env" } }]
            initContainers = [{
              name            = "nat"
              image           = "{{ .Values.image.repository }}:{{ .Values.image.tag }}"
              securityContext = { capabilities = { add = ["NET_ADMIN"] } }
              command = ["sh", "-ec", <<-EOF
                iptables -t nat -C POSTROUTING -o eth0 -j MASQUERADE 2>/dev/null ||
                  iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE
              EOF
              ]
            }]
            service = {
              wireguard = {
                annotations = {
                  "external-dns.kubernetes.io/hostname" = "wg.${var.domain}"
                  "external-dns.kubernetes.io/target"   = "public.${var.domain}"
                }
              }
            }
            extraDeploy = [{
              apiVersion = "v1"
              kind       = "Secret"
              metadata = {
                name        = "wireguard-env"
                labels      = { "homelab.khuedoan.com/bao-secret" = "true" }
                annotations = { "argocd.argoproj.io/sync-wave" = "-1" }
              }
              stringData = {
                OIDC_CLIENT_SECRET           = "bao:secret/data/wireguard/sso#client_secret"
                WG_PORTAL_WEB_SESSION_SECRET = "bao:secret/data/wireguard/web#session_secret"
                WG_PORTAL_WEB_CSRF_SECRET    = "bao:secret/data/wireguard/web#csrf_secret"
              }
              }, {
              apiVersion = "gateway.networking.k8s.io/v1"
              kind       = "HTTPRoute"
              metadata   = { name = "wireguard" }
              spec = {
                parentRefs = [{ name = "gateway", namespace = "istio-system", sectionName = "https" }]
                hostnames  = ["vpn.${var.domain}"]
                rules      = [{ backendRefs = [{ name = "wireguard", port = 8888 }] }]
              }
            }]
            config = {
              core = {
                disable_admin_user           = true
                create_default_peer_on_login = true
                self_provisioning_allowed    = true
              }
              advanced = { use_ip_v6 = false } # TODO enable IPv6
              web      = { external_url = "https://vpn.${var.domain}" }
              auth = {
                hide_login_form = true
                oidc = [{
                  provider_name        = "sso"
                  display_name         = "SSO"
                  base_url             = "https://auth.${var.domain}/oauth2/openid/wireguard"
                  client_id            = "wireguard"
                  client_secret        = "$${OIDC_CLIENT_SECRET}"
                  extra_scopes         = ["profile", "email", "groups"]
                  registration_enabled = true
                  field_map = {
                    firstname   = "preferred_username"
                    user_groups = "groups"
                  }
                  admin_mapping = { admin_group_regex = "^admins@auth\\.${replace(var.domain, ".", "\\.")}$" }
                }]
              }
            }
          })
        }
      }]
      destination       = { name = "in-cluster", namespace = "wireguard" }
      ignoreDifferences = local.ignore_differences
      syncPolicy        = local.sync_policy
    }
  })
  lifecycle {
    prevent_destroy = true
  }
}
