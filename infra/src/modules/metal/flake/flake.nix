{
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-26.05";
    disko = {
      url = "github:nix-community/disko";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    nixie = {
      url = "github:khuedoan/nixie";
      inputs.nixpkgs.follows = "nixpkgs";
    };
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
        hosts = builtins.fromJSON (builtins.readFile ./hosts.json);
        cluster = builtins.fromJSON (builtins.readFile ./cluster.json);
      };
    };
}
