{ modulesPath, ... }:

{
  imports = [
    (modulesPath + "/installer/netboot/netboot-minimal.nix")
  ];

  installer.cloneConfig = false;

  users.users.root.openssh.authorizedKeys.keys = [
    "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIN5ue4np7cF34f6dwqH1262fPjkowHQ8irfjVC156PCG"
    "ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBM/WQcPFuzsPmfXSM1GGkIndFcDRirTl5Aqsou8lWPJyUNZOdFt2cWlUkm+Q1F+LFJQ2+YdIXPZlTqhWLF1eWuY= khuedoan@codeserver"
  ];

  system.stateVersion = "25.05";
}
