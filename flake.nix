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
    in
    {

      formatter.${system} = pkgs.nixfmt-tree;

      devShells.${system}.default = pkgs.mkShell {
        packages = with pkgs; [
          nixie.packages.${system}.default
          (python3.withPackages (ps: [ ps.pyyaml ]))
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
