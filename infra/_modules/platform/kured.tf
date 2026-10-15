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
        targetRevision = "6.1.0"
        helm = {
          releaseName = "kured"
          values = yamlencode({
            configuration = {
              annotateNodes             = true
              useRebootSentinelHostPath = false
              rebootSentinelCommand     = <<-EOT
                /run/current-system/sw/bin/sh -c '
                  for file in kernel initrd kernel-modules; do
                    booted=$(/run/current-system/sw/bin/readlink -f /run/booted-system/$file)
                    configured=$(/run/current-system/sw/bin/readlink -f /nix/var/nix/profiles/system/$file)
                    test "$booted" = "$configured" || exit 0
                  done
                  exit 1
                '
              EOT
              rebootCommand             = "/run/current-system/sw/bin/systemctl reboot"
              timeZone                  = "Asia/Ho_Chi_Minh"
              startTime                 = "03:00"
              endTime                   = "04:00"
              drainTimeout              = "10m"
            }
            resources = {
              requests = { cpu = "10m", memory = "32Mi" }
              limits   = { memory = "128Mi" }
            }
            metrics = { create = true }
            service = { create = true }
          })
        }
      }]
      destination       = { name = "in-cluster", namespace = "kured" }
      ignoreDifferences = local.ignore_differences
      syncPolicy        = local.sync_policy
    }
  })
  lifecycle {
    prevent_destroy = true
  }
}
