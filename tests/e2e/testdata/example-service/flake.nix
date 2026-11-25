{
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-26.05";
  };

  outputs = { self, nixpkgs, ... }: {
    packages = builtins.mapAttrs (
      _system: pkgs:
      let
        revision = self.rev or "development";

        service = pkgs.buildGoModule {
          pname = "example-service";
          version = "0.1.0";

          src = ./.;
          vendorHash = null;
          env.CGO_ENABLED = "0";
          ldflags = [ "-X=main.version=${revision}" ];

          preConfigure = ''
            export XDG_CONFIG_HOME="$TMPDIR"
          '';
        };
      in
      {
        inherit (pkgs) skopeo;

        image = pkgs.dockerTools.buildLayeredImage {
          name = service.pname;
          tag = revision;
          config.Entrypoint = [ "${service}/bin/${service.pname}" ];
        };
      }
    ) nixpkgs.legacyPackages;
  };
}
