{ lib, modulesPath, ... }:

{
  imports = [
    (modulesPath + "/installer/netboot/netboot-minimal.nix")
  ];

  installer.cloneConfig = false;

  # Override Nixie's forced SSH disable for installer diagnostics.
  services.openssh.enable = lib.mkOverride 40 true;
  # Keep port 22 reserved for Nixie's installed-system check.
  services.openssh.ports = [ 2222 ];
  services.openssh.settings.PasswordAuthentication = false;
  services.openssh.settings.KbdInteractiveAuthentication = false;
  users.users.root.openssh.authorizedKeys.keyFiles = [ ../authorized_keys ];

  system.stateVersion = "25.05";
}
