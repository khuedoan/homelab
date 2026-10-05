resource "kubectl_manifest" "kured" {
  server_side_apply = true
  yaml_body = yamlencode({
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = "kured"
      namespace = "argocd"
      labels    = { "app.kubernetes.io/part-of" = "homelab" }
    }
    spec = {
      project = "default"
      sources = [{
        repoURL        = "https://kubereboot.github.io/charts"
        chart          = "kured"
        targetRevision = "4.7.0"
        helm = {
          releaseName = "kured"
          values = yamlencode({
            # TODO: Revisit resource sizing after measuring usage.
            resources = { requests = { cpu = "10m", memory = "32Mi" } }
            configuration = {
              annotateNodes = true
              rebootCommand = "/run/current-system/sw/bin/systemctl reboot"
              rebootSentinelCommand = join(" ", [
                "/bin/sh -c 'for part in kernel initrd kernel-modules; do",
                "booted=$(/run/current-system/sw/bin/readlink \"/run/booted-system/$part\") || exit 2;",
                "desired=$(/run/current-system/sw/bin/readlink \"/nix/var/nix/profiles/system/$part\") || exit 2;",
                "[ \"$booted\" = \"$desired\" ] || exit 0;",
                "done; exit 1'",
              ])
              timeZone = "Asia/Ho_Chi_Minh"
            }
          })
        }
      }]
      destination       = { server = "https://kubernetes.default.svc", namespace = "kured" }
      ignoreDifferences = local.ignore_differences
      syncPolicy        = local.sync_policy
    }
  })
  lifecycle {
    prevent_destroy = true
  }
}
