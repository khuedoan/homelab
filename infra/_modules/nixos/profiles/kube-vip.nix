{ k3s, lib, ... }:

let
  metadata = {
    name = "kube-vip";
    namespace = "kube-system";
  };
in
{
  services.k3s.manifests.kube-vip.content = [
    {
      apiVersion = "v1";
      kind = "ServiceAccount";
      inherit metadata;
    }
    {
      apiVersion = "rbac.authorization.k8s.io/v1";
      kind = "Role";
      inherit metadata;
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
      ];
    }
    {
      apiVersion = "rbac.authorization.k8s.io/v1";
      kind = "RoleBinding";
      inherit metadata;
      roleRef = {
        apiGroup = "rbac.authorization.k8s.io";
        kind = "Role";
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
            dnsPolicy = "ClusterFirstWithHostNet";
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
                env =
                  lib.mapAttrsToList (name: value: { inherit name value; }) {
                    address = k3s.vip;
                    vip_arp = "true";
                    cp_enable = "true";
                    cp_namespace = "kube-system";
                    vip_leaderelection = "true";
                    vip_leasename = "plndr-cp-lock";
                    svc_enable = "false";
                    lb_enable = "false";
                    enable_node_labeling = "false";
                    KUBERNETES_SERVICE_HOST = "127.0.0.1";
                    KUBERNETES_SERVICE_PORT = "6443";
                  }
                  ++ [
                    {
                      name = "vip_nodename";
                      valueFrom.fieldRef.fieldPath = "spec.nodeName";
                    }
                  ];
              }
            ];
          };
        };
      };
    }
  ];
}
