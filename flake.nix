{
  inputs = {
    nixpkgs = {
      url = "github:nixos/nixpkgs/nixos-26.05";
    };
    disko = {
      url = "github:nix-community/disko";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    nixie = {
      url = "github:khuedoan/nixie";
    };
  };

  outputs =
    {
      nixpkgs,
      disko,
      nixie,
      ...
    }:
    let
      system = "x86_64-linux";

      pkgs = import nixpkgs { inherit system; };
    in
    {
      devShells.${system}.default = pkgs.mkShell {
        packages = with pkgs; [
          dyff
          gnumake
          go
          gotestsum
          kubectl
          kubernetes-helm
          nixfmt-tree
          nixos-anywhere
          nixos-rebuild
          openssh
          opentofu
        ] ++ [
          nixie.packages.${system}.default
        ];
      };

      nixosConfigurations = import ./metal {
        inherit nixpkgs disko nixie;
      };
    };
}
