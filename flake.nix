{
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-26.05";
    nixie.url = "github:khuedoan/nixie";
  };

  outputs =
    {
      nixpkgs,
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
        vendorHash = "sha256-E9k3y+b/cz35T6athMccHnWV+DCWXDvFKukSmnxMZCE=";
        preCheck = ''
          export SFTP_SERVER="${pkgs.openssh}/libexec/sftp-server"
        '';
      };
    in
    {
      packages.${system}.toolbox = toolbox;

      devShells.${system}.default = pkgs.mkShell {
        SFTP_SERVER = "${pkgs.openssh}/libexec/sftp-server";
        packages =
          with pkgs;
          [
            gnumake
            go
            golangci-lint
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
    };
}
