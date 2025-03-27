{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-24.11";
  };

  outputs = { self, nixpkgs }:
  let
    supportedSystems = nixpkgs.lib.genAttrs [
      "x86_64-linux"
    ];
  in
  {
    devShells = supportedSystems (system: {
      default = with nixpkgs.legacyPackages.${system}; mkShell {
        packages = [
          go
          go-task
          kubectl
          opentofu
        ];
      };
    });
  };
}
