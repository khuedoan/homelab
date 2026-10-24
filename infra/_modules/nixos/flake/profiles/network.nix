{ hostConfig, network, ... }:

{
  networking = {
    useDHCP = false;
    useNetworkd = true;
  };

  systemd.network.networks."10-uplink" = {
    matchConfig.MACAddress = hostConfig.mac_address;
    address = [ "${hostConfig.ip}/${toString network.prefix_length}" ];
    gateway = [ network.gateway ];
    dns = [
      "8.8.8.8"
      "1.1.1.1"
    ];
    networkConfig = {
      DHCP = "ipv6";
      IPv6AcceptRA = true;
      KeepConfiguration = "static";
    };
  };
}
