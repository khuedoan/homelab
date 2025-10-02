{ nixpkgs, disko }:

{
  installer = nixpkgs.lib.nixosSystem {
    system = "x86_64-linux";
    modules = [
      ./installer.nix
    ];
  };
  metal1 = nixpkgs.lib.nixosSystem {
    system = "x86_64-linux";
    modules = [
      disko.nixosModules.disko
      ./configuration.nix
      {
        networking.hostName = "metal1";
      }
    ];
  };
  metal2 = nixpkgs.lib.nixosSystem {
    system = "x86_64-linux";
    modules = [
      disko.nixosModules.disko
      ./configuration.nix
      {
        networking.hostName = "metal2";
      }
    ];
  };
}
