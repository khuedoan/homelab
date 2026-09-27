{
  nixpkgs,
  disko,
  nixie,
}:

let
  hosts = builtins.fromJSON (builtins.readFile ../metal/hosts.json);
in
{
  installer = nixpkgs.lib.nixosSystem {
    system = "x86_64-linux";
    modules = [
      nixie.nixosModules.nixie-agent
      ./installer.nix
    ];
  };
}
// nixpkgs.lib.mapAttrs (
  name: _:
  nixpkgs.lib.nixosSystem {
    system = "x86_64-linux";
    modules = [
      disko.nixosModules.disko
      ./configuration.nix
      {
        networking.hostName = name;
      }
    ];
  }
) hosts
