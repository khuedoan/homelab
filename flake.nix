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
            ansible
            ansible-lint
            bmake
            diffutils
            docker
            docker-compose
            dyff
            git
            glibcLocales
            go
            gotestsum
            iproute2
            jq
            k9s
            kanidm
            kube3d
            kubectl
            kubernetes-helm
            kustomize
            libisoburn
            neovim
            nixfmt-tree
            nixos-anywhere
            openssh_gssapi
            opentofu # Drop-in replacement for Terraform
            p7zip
            pre-commit
            qrencode
            shellcheck
            wireguard-tools
            yamllint

            (python3.withPackages (
              p: with p; [
                jinja2
                kubernetes
                mkdocs-material
                netaddr
                pexpect
                rich
              ]
            ))
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
