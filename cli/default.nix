{
  lib,
  buildGoModule,
  git,
  openssh,
}:

buildGoModule {
  pname = "lab";
  version = "0.1.0";
  src = lib.cleanSource ./.;
  vendorHash = "sha256-bxxqUoFcmMxXXNM43SOu9dmMVLCj/uWkWtmJymTIVaY=";
  subPackages = [ "cmd/lab" ];
  nativeCheckInputs = [ git ];
  SFTP_SERVER = "${openssh}/libexec/sftp-server";
  checkPhase = ''
    runHook preCheck
    go test ./...
    runHook postCheck
  '';
  meta.mainProgram = "lab";
}
