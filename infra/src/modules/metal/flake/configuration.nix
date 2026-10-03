{ modulesPath, ... }:

{
  imports = [
    (modulesPath + "/profiles/all-hardware.nix")
    ./disks.nix
  ];

  boot.loader.systemd-boot.enable = true;
  boot.loader.efi.canTouchEfiVariables = true;

  nix.settings = {
    experimental-features = [
      "nix-command"
      "flakes"
    ];
    trusted-users = [ "admin" ];
  };
  nix.optimise.automatic = true;
  nix.gc = {
    automatic = true;
    dates = "weekly";
    options = "--delete-older-than 30d";
  };

  services.openssh.enable = true;
  security.sudo.wheelNeedsPassword = false;

  users.users.admin = {
    isNormalUser = true;
    extraGroups = [
      "wheel"
    ];
    openssh.authorizedKeys.keyFiles = [ ./authorized_keys ];
  };
  users.users.root.openssh.authorizedKeys.keyFiles = [ ./authorized_keys ];

  system.stateVersion = "25.05";
}
