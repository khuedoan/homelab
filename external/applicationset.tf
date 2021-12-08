resource "kubernetes_manifest" "external_applicationset" {
  manifest = {
    apiVersion = "argoproj.io/v1alpha1"
    kind = "ApplicationSet"
    metadata = {
      name = "external"
      namespace = "argocd"
    }
    spec = {
      generators = [
        {
          git = {
            directories = [
              {
                path = "external/*"
              }
            ]
            repoURL = "https://github.com/khuedoan/homelab.git"
            revision = "master"
          }
        },
      ]
      template = {
        metadata = {
          name = "{{path.basename}}"
        }
        spec = {
          destination = {
            name = "in-cluster"
            namespace = "{{path.basename}}"
          }
          project = "default"
          source = {
            path = "{{path}}"
            repoURL = "https://github.com/khuedoan/homelab.git"
          }
          syncPolicy = {
            automated = {
              prune = true
              selfHeal = true
            }
          }
        }
      }
    }
  }
}
