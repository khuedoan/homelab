{
  inputs = {
    nixpkgs = {
      url = "github:nixos/nixpkgs/nixos-25.05";
    };
    flake-utils = {
      url = "github:numtide/flake-utils";
    };
    disko = {
      url = "github:nix-community/disko";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
      disko,
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs { inherit system; };
      in
      {
        devShells.default = pkgs.mkShell {
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
            openssh_gssapi
            opentofu
          ];
        };
      }
    )
    // {
      nixosConfigurations = import ./metal {
        inherit nixpkgs disko;
      };
      nixosPxeServer =
        let
          installer = (import ./metal { inherit nixpkgs disko; }).installer;
          hostPkgs = installer.pkgs;
          build = installer.config.system.build;
        in
        hostPkgs.writeShellApplication {
          name = "nixos-pxe-server";
          # TODO Pixiecore is unmaintained, probably need to find a new one
          text = ''
            exec ${hostPkgs.pixiecore}/bin/pixiecore \
              boot \
              ${build.kernel}/bzImage \
              ${build.netbootRamdisk}/initrd \
              --cmdline "init=${build.toplevel}/init loglevel=4" \
              --dhcp-no-bind \
              --debug \
              --port 8080
          '';
        };
    };
}
