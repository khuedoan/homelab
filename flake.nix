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
    in
    {
      packages.${system}.cf = cf;

      formatter.${system} = pkgs.nixfmt-tree;

      devShells.${system}.default = pkgs.mkShell {
        packages = with pkgs; [
          cf
          nixie.packages.${system}.default
          (python3.withPackages (ps: [ ps.pyyaml ]))
          jq
          kubectl
          kubernetes-helm
          nixfmt-tree
          nixos-anywhere
          nixos-rebuild
          openssh
          opentofu
          terragrunt
        ];
      };
    };
}
