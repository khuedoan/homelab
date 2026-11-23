{
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-26.05";
  };

  outputs = { self, nixpkgs, ... }: {
    packages = builtins.mapAttrs (
      _system: pkgs:
      let
        version = self.rev or "development";

        service = pkgs.buildGoModule {
          pname = "example-service";
          version = "0.1.0";

          src = nixpkgs.lib.fileset.toSource {
            root = ./.;
            fileset = nixpkgs.lib.fileset.unions [
              ./go.mod
              ./main.go
            ];
          };

          vendorHash = null;
          env.CGO_ENABLED = "0";
          ldflags = [ "-X=main.version=${version}" ];

          preConfigure = ''
            export XDG_CONFIG_HOME="$TMPDIR"
          '';
        };
      in
      {
        inherit (pkgs) skopeo;

        image = pkgs.dockerTools.buildLayeredImage {
          name = "example-service";
          tag = version;
          config.Entrypoint = [ "${service}/bin/example-service" ];
        };
      }
    ) nixpkgs.legacyPackages;
  };
}
