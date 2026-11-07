{
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-26.05";
    nixpkgs-unstable.url = "github:nixos/nixpkgs/nixos-unstable";
    # TODO switch nixie back to master or release (testing out a new refactor)
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
      inherit (nixpkgs-unstable.legacyPackages.${system}) mdbook mdbook-mermaid;
    in
    {
      formatter.${system} = pkgs.nixfmt-tree;

      devShells.${system}.default = pkgs.mkShell {
        packages = with pkgs; [
          nixie.packages.${system}.default
          just
          (python3.withPackages (ps: [
            ps.cloudflare
            ps.hvac
            ps.kanidm
            ps.pyyaml
            ps.requests
          ]))
          go
          golangci-lint
          gotestsum
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
          pass
          terragrunt
          velero
        ];
      };
    };
}
