{ pkgs, modulesPath, ... }:

{
  imports = [
    (modulesPath + "/installer/netboot/netboot-minimal.nix")
  ];

  networking.hostName = "nixos-installer";
  services.openssh.enable = true;

  users.users.root.openssh.authorizedKeys.keys = [
    "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIN5ue4np7cF34f6dwqH1262fPjkowHQ8irfjVC156PCG"
  ];

  environment = {
    systemPackages = with pkgs; [
      btop
    ];
  };

  systemd.services."phone-home" =
    let
      reportScript = pkgs.writeShellScript "phone-home.sh" ''
        set -euo pipefail

        sleep 30

        iface=$(${pkgs.iproute2}/bin/ip route show default | ${pkgs.coreutils}/bin/cut -d ' ' -f 5)
        mac=$(cat /sys/class/net/$iface/address)
        ip=$(${pkgs.iproute2}/bin/ip route show default | ${pkgs.coreutils}/bin/cut -d ' ' -f 9)

        ${pkgs.curl}/bin/curl -sf -X POST "http://192.168.1.15:5000/report" \
          --data-urlencode "mac=$mac" \
          --data-urlencode "ip=$ip"
      '';
    in
    {
      description = "Report IP address after PXE boot";
      after = [ "network-online.target" ];
      wants = [ "network-online.target" ];
      serviceConfig = {
        Type = "oneshot";
        ExecStart = [ reportScript ];
      };
      wantedBy = [ "multi-user.target" ];
    };
}
