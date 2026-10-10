{
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-26.05";
    disko = {
      url = "github:nix-community/disko";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    nixie.url = "github:khuedoan/nixie";
  };

  outputs =
    {
      nixpkgs,
      disko,
      nixie,
      ...
    }:
    {
      nixosConfigurations = import ./. {
        inherit nixpkgs disko nixie;
        hosts =
          if builtins.pathExists ./hosts.json then
            builtins.fromJSON (builtins.readFile ./hosts.json)
          else
            { };
        cluster =
          if builtins.pathExists ./cluster.json then
            builtins.fromJSON (builtins.readFile ./cluster.json)
          else
            null;
      };
    };
}
