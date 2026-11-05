{
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-26.05";
    nixpkgs-unstable.url = "github:nixos/nixpkgs/nixos-unstable";
    nixie.url = "github:khuedoan/nixie/pull-based-install";
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
    in
    {
      packages.${system}.cf = cf;

      formatter.${system} = pkgs.nixfmt-tree;

      devShells.${system}.default = pkgs.mkShell {
        packages = with pkgs; [
          cf
          nixie.packages.${system}.default
          just
          (python3.withPackages (ps: [
            ps.hvac
            ps.kanidm
            ps.pyyaml
            ps.requests
          ]))
          jq
          kanidm_1_11
          kubectl
          kubernetes-helm
          mdbook
          mdbook-mermaid
          nixfmt-tree
          ruff
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
