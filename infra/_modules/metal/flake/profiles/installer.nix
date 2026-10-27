{ modulesPath, ... }:

{
  imports = [
    (modulesPath + "/installer/netboot/netboot-minimal.nix")
  ];

  installer.cloneConfig = false;

  users.users.root.openssh.authorizedKeys.keyFiles = [ ../authorized_keys ];

  system.stateVersion = "25.05";
}
