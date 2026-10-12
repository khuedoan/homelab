# Offline regression: nix-instantiate --eval --strict ./tests/node-ip.nix
# Only inspect extraFlags, without evaluating nixpkgs or building a machine.
let
  flagsFor = hostConfig:
    (import ../profiles/k3s-server.nix {
      inherit hostConfig;
      config = { };
      lib.optionals = condition: values: if condition then values else [ ];
      pkgs = { };
      k3s = { clusterInit = true; vip = "192.0.2.100"; };
    }).services.k3s.extraFlags;
  initial = flagsFor { mac_address = "00:11:22:33:44:55"; };
  discovered = flagsFor { ip = "192.0.2.10"; };
  empty = flagsFor { ip = ""; };
  unset = flagsFor { ip = null; };
  hasNodeIP = flags: builtins.any (flag: builtins.match "--node-ip=.*" flag != null) flags;
in
assert !hasNodeIP initial;
assert !hasNodeIP empty;
assert !hasNodeIP unset;
assert builtins.elem "--node-ip=192.0.2.10" discovered;
assert builtins.elem "--tls-san=192.0.2.100" initial;
true
