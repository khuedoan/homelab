{
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-26.05";
    nixpkgs-unstable.url = "github:nixos/nixpkgs/nixos-unstable";
    nixie.url = "github:khuedoan/nixie";
  };

  outputs =
    {
      nixpkgs,
      nixpkgs-unstable,
      nixie,
      ...
    }:
    let
      system = "x86_64-linux";

      pkgs = nixpkgs.legacyPackages.${system};
      # TODO use cf on nixpkgs stable
      inherit (nixpkgs-unstable.legacyPackages.${system}) cf;
      sftpServer = "${pkgs.openssh}/libexec/sftp-server";

      lab = pkgs.buildGoModule {
        pname = "lab";
        version = "0.1.0";
        src = pkgs.lib.cleanSource ./cli;
        vendorHash = "sha256-bxxqUoFcmMxXXNM43SOu9dmMVLCj/uWkWtmJymTIVaY=";
        subPackages = [ "cmd/lab" ];
        nativeCheckInputs = [ pkgs.git ];
        SFTP_SERVER = sftpServer;
        meta.mainProgram = "lab";
      };
    in
    {
      packages.${system} = {
        inherit lab cf;
        default = lab;
      };

      checks.${system}.lab = lab;
      formatter.${system} = pkgs.nixfmt-tree;

      devShells.${system}.default = pkgs.mkShell {
        SFTP_SERVER = sftpServer;
        packages = with pkgs; [
          cf
          lab
          nixie.packages.${system}.default
          gnumake
          go
          golangci-lint
          gotestsum
          jq
          kubectl
          kubernetes-helm
          nixfmt-tree
          nixos-anywhere
          nixos-rebuild
          openssh
          opentofu
          terragrunt
          velero
          (python3.withPackages (pythonPackages: [
            pythonPackages.mkdocs-material
          ]))
        ];
      };
    };
}
