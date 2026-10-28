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
      inherit (nixpkgs-unstable.legacyPackages.${system}) cf mdbook mdbook-mermaid;
      lab = pkgs.callPackage ./cli { };
    in
    {
      packages.${system} = {
        inherit lab cf;
        default = lab;
      };

      checks.${system}.lab = lab;
      formatter.${system} = pkgs.nixfmt-tree;

      devShells.${system}.default = pkgs.mkShell {
        inherit (lab) SFTP_SERVER;
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
          mdbook
          mdbook-mermaid
          nixfmt-tree
          nixos-anywhere
          nixos-rebuild
          openssh
          opentofu
          terragrunt
          velero
        ];
      };
    };
}
