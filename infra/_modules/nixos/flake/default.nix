{
  nixpkgs,
  disko,
  nixie,
  hosts ? { },
  cluster ? null,
}:

let
  initHost =
    if builtins.hasAttr cluster.init_host hosts then
      cluster.init_host
    else
      throw "Cluster init_host '${cluster.init_host}' is not present in hosts.json";
in
{
  installer = nixpkgs.lib.nixosSystem {
    system = "x86_64-linux";
    modules = [
      nixie.nixosModules.nixie-agent
      ./profiles/installer.nix
    ];
  };
}
// nixpkgs.lib.mapAttrs (
  name: hostConfig:
  assert builtins.seq initHost true;
  nixpkgs.lib.nixosSystem {
    system = "x86_64-linux";
    specialArgs = {
      inherit hostConfig;
      network = cluster.network;
      k3s = {
        clusterInit = name == initHost;
        vip = cluster.vip;
        loadBalancerCIDR = cluster.load_balancer_cidr;
        nodeIPs = nixpkgs.lib.mapAttrsToList (_: host: host.ip) hosts;
      };
    };
    modules = [
      disko.nixosModules.disko
      ./configuration.nix
      ./profiles/network.nix
      ./profiles/k3s-server.nix
      ./profiles/kube-vip.nix
      ./profiles/rook-ceph.nix
      {
        networking.hostName = name;
      }
    ];
  }
) hosts
