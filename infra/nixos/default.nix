{
  nixpkgs,
  disko,
  nixie,
}:

let
  hosts = builtins.fromJSON (builtins.readFile ./hosts.json);
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
  nixpkgs.lib.nixosSystem {
    system = "x86_64-linux";
    specialArgs = {
      inherit hostConfig;
    };
    modules = [
      disko.nixosModules.disko
      ./configuration.nix
      {
        networking.hostName = name;
      }
    ];
  }
) hosts
