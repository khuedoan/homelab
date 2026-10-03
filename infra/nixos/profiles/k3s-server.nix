{
  config,
  lib,
  k3s,
  ...
}:

{
  assertions = [
    {
      assertion =
        config.services.k3s.clusterInit
        || (config.services.k3s.serverAddr != "" && config.services.k3s.tokenFile != null);
      message = "Non-initializer k3s servers require serverAddr and tokenFile.";
    }
  ];

  # k3s requires these to route traffic between pods and services.
  boot.kernelModules = [ "br_netfilter" ];
  boot.kernel.sysctl = {
    "net.ipv4.ip_forward" = 1;
    "net.bridge.bridge-nf-call-iptables" = 1;
  };

  networking.firewall = {
    # https://docs.k3s.io/installation/requirements#inbound-rules-for-k3s-nodes
    allowedTCPPorts = [
      6443 # Kubernetes API server
      10250 # kubelet
      2379 # etcd client
      2380 # etcd peer
      80 # HTTP ingress
      443 # HTTPS ingress
    ];
    allowedUDPPorts = [
      8472 # flannel VXLAN
    ];
  };

  systemd.services.k3s.unitConfig = lib.mkIf (!k3s.clusterInit) {
    ConditionPathExists = config.services.k3s.tokenFile;
    RequiresMountsFor = "/var/lib/rancher/k3s";
  };

  services.k3s = {
    enable = true;
    role = "server";
    # Joiners remain stopped until toolbox enrolls them with the secure token.
    clusterInit = k3s.clusterInit;
    serverAddr = if k3s.clusterInit then "" else "https://${k3s.vip}:6443";
    tokenFile = if k3s.clusterInit then null else "/var/lib/rancher/k3s/enrollment/token";
    disable = [
      "local-storage" # storage is provided by Rook Ceph
      "traefik" # ingress is provided by NGINX
    ];
    extraFlags = [
      "--tls-san=${k3s.vip}"
      "--cluster-cidr=10.42.0.0/16"
      "--service-cidr=10.43.0.0/16"
    ];
  };
}
