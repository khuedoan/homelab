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
      toolbox = pkgs.buildGoModule {
        pname = "toolbox";
        version = "0.1.0";
        src = builtins.path {
          path = ./toolbox;
          name = "toolbox-src";
        };
        vendorHash = "sha256-BDyg4x042o2XcYRdbkNpX3q3xHssbtgYNTv/LE4zvoQ=";
        nativeCheckInputs = [ pkgs.git ];
      };
    in
    {
      packages.${system}.toolbox = toolbox;

      devShells.${system}.default = pkgs.mkShell {
        packages =
          with pkgs;
          [
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
            terragrunt
            (python3.withPackages (pythonPackages: [
              pythonPackages.mkdocs-material
            ]))
          ]
          ++ [
            nixie.packages.${system}.default
            toolbox
          ];
      };

      nixosConfigurations = import ./infra/nixos {
        inherit nixpkgs disko nixie;
      };
    };
}
