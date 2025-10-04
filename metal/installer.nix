{
  pkgs,
  lib,
  config,
  modulesPath,
  ...
}:

{
  imports = [
    (modulesPath + "/installer/netboot/netboot-minimal.nix")
  ];

  networking.hostName = "nixos-installer";
  services.openssh.enable = true;

  # Set root password to "nixos-installer" for initial SSH access
  users.users.root = {
    password = "nixos-installer";
    initialHashedPassword = lib.mkForce null;
    openssh.authorizedKeys.keys = [
      "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIN5ue4np7cF34f6dwqH1262fPjkowHQ8irfjVC156PCG"
    ];
  };

  environment = {
    systemPackages = with pkgs; [
      btop
    ];
  };

  systemd.services."installer-callback" =
    let
      reportScript = pkgs.writeShellScript "installer-callback.sh" ''
        set -euo pipefail

        # Wait for default route and IP address to be ready
        echo "Waiting for network to be ready..."
        while true; do
          if ${pkgs.iproute2}/bin/ip route show default > /dev/null 2>&1; then
            iface=$(${pkgs.iproute2}/bin/ip route show default | ${pkgs.coreutils}/bin/cut -d ' ' -f 5)
            ip=$(${pkgs.iproute2}/bin/ip route show default | ${pkgs.coreutils}/bin/cut -d ' ' -f 9)
            if [ -n "$iface" ] && [ -n "$ip" ] && [ "$ip" != "0.0.0.0" ]; then
              echo "Network is ready: interface=$iface ip=$ip"
              break
            fi
          fi
          sleep 1
        done

        mac=$(cat /sys/class/net/$iface/address)

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

  # TODO
  # system.stateVersion = config.system.nixos.version;
}
