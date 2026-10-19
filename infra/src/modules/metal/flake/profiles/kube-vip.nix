{
  k3s,
  lib,
  pkgs,
  ...
}:

let
  metadata = {
    name = "kube-vip";
    namespace = "kube-system";
  };
in
{
  services.k3s.manifests.kube-vip-cloud-controller.source = pkgs.fetchurl {
    url = "https://raw.githubusercontent.com/kube-vip/kube-vip-cloud-provider/v0.0.13/manifest/kube-vip-cloud-controller.yaml";
    hash = "sha256-rHjtJsI5TalHFfJM/yDg1DaWmFoK0vXhaHdxmF3aNoA=";
  };
  services.k3s.manifests.kube-vip-pool.content = {
    apiVersion = "v1";
    kind = "ConfigMap";
    metadata = {
      name = "kubevip";
      namespace = "kube-system";
    };
    data.cidr-global = k3s.loadBalancerCIDR;
  };
  services.k3s.manifests.kube-vip.content = [
    {
      apiVersion = "v1";
      kind = "ServiceAccount";
      inherit metadata;
    }
    {
      apiVersion = "rbac.authorization.k8s.io/v1";
      kind = "ClusterRole";
      metadata.name = "kube-vip";
      rules = [
        {
          apiGroups = [ "coordination.k8s.io" ];
          resources = [ "leases" ];
          verbs = [
            "get"
            "create"
            "update"
          ];
        }
        {
          apiGroups = [ "" ];
          resources = [ "services" ];
          verbs = [
            "get"
            "list"
            "watch"
            "update"
          ];
        }
        {
          apiGroups = [ "" ];
          resources = [ "services/status" ];
          verbs = [ "update" ];
        }
        {
          apiGroups = [ "discovery.k8s.io" ];
          resources = [ "endpointslices" ];
          verbs = [
            "get"
            "list"
            "watch"
          ];
        }
      ];
    }
    {
      apiVersion = "rbac.authorization.k8s.io/v1";
      kind = "ClusterRoleBinding";
      metadata.name = "kube-vip";
      roleRef = {
        apiGroup = "rbac.authorization.k8s.io";
        kind = "ClusterRole";
        name = "kube-vip";
      };
      subjects = [
        {
          kind = "ServiceAccount";
          name = "kube-vip";
          namespace = "kube-system";
        }
      ];
    }
    {
      apiVersion = "apps/v1";
      kind = "DaemonSet";
      inherit metadata;
      spec = {
        selector.matchLabels.app = "kube-vip";
        template = {
          metadata.labels.app = "kube-vip";
          spec = {
            serviceAccountName = "kube-vip";
            hostNetwork = true;
            affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms = [
              {
                matchExpressions = [
                  {
                    key = "node-role.kubernetes.io/control-plane";
                    operator = "Exists";
                  }
                ];
              }
            ];
            tolerations = [
              {
                operator = "Exists";
                effect = "NoSchedule";
              }
              {
                operator = "Exists";
                effect = "NoExecute";
              }
            ];
            containers = [
              {
                name = "kube-vip";
                image = "ghcr.io/kube-vip/kube-vip:v1.2.4";
                args = [ "manager" ];
                securityContext.capabilities = {
                  add = [
                    "NET_ADMIN"
                    "NET_RAW"
                  ];
                  drop = [ "ALL" ];
                };
                env = lib.mapAttrsToList (name: value: { inherit name value; }) {
                  address = k3s.vip;
                  vip_arp = "true";
                  cp_enable = "true";
                  vip_leaderelection = "true";
                  svc_enable = "true";
                  svc_election = "true";
                  KUBERNETES_SERVICE_HOST = "127.0.0.1";
                  KUBERNETES_SERVICE_PORT = "6443";
                };
              }
            ];
          };
        };
      };
    }
  ];
}
