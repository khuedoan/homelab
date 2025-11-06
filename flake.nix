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
    (flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs { inherit system; };
        installer = (import ./metal { inherit nixpkgs disko; }).installer;
        build = installer.config.system.build;
        homelabInstall = pkgs.buildGoModule {
          pname = "homelab-install";
          version = "0.1.0";
          src = ./tools;
          # TODO better way to build this shit
          vendorHash = "sha256-rTJt3UWRUyhRDx1Sdno0fFBYMb4RPtzB7Z7sg45ZJ8o=";

          postInstall = ''
            wrapProgram $out/bin/homelab-install \
              --add-flags "-kernel ${build.kernel}/bzImage" \
              --add-flags "-initrd ${build.netbootRamdisk}/initrd" \
              --add-flags "-init ${build.toplevel}/init" \
              --prefix PATH : ${pkgs.lib.makeBinPath [ pkgs.nixos-anywhere ]}
          '';

          nativeBuildInputs = [ pkgs.makeWrapper ];
        };
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
            openssh
            opentofu
          ];
        };

        packages = {
          homelabInstall = homelabInstall;
          nixosPxeServer = homelabInstall; # Alias for backwards compatibility
          default = homelabInstall;
        };
      }
    ))
    // {
      nixosConfigurations = import ./metal {
        inherit nixpkgs disko;
      };
    };
}
