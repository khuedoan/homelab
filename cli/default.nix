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
  vendorHash = "sha256-P0QuqcH86h4JB7KFafMVh8d30q3o/3pnNdEMfvm2TRU=";
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
