# Prerequisites

## BIOS setup

> You need to do it once per machine if the default config is not sufficent,
> usually for consumer hardware this can not be automated
> (it requires something like [IPMI](https://en.wikipedia.org/wiki/Intelligent_Platform_Management_Interface) to automate).

Below is my BIOS setup, your motherboard may have a different name for the options, so you'll need to adapt it with your hardware.

```json
{
    "Devices": {
        "Network Setup": {
            "PXE IPv4": true,
            "PXE IPv6": false
        }
    },
    "Advanced": {
        "CPU Setup": {
            "VT-d": true
        }
    },
    "Power": {
        "Automatic Power On": {
            "WoL": "Automatic"
        }
    },
    "Security": {
        "Secure Boot": false
    },
    "Startup": {
        "CSM": false
    }
}
```

## Update config files

Gather the following information:

- [ ] MAC address for each machine
- [ ] OS disk name (for example `/dev/sda`)
- [ ] Network interface name (for example `eth0`)
- [ ] Choose a static IP address for each machine (just the desired address, we don't set anything up yet)

[Fork this repo](https://github.com/khuedoan/homelab) and update these config files based on those information:

- [ ] `metal/hosts.yml`
- [ ] `metal/group_vars/all.yml`
- [ ] TODO git config in `bootstrap/...`
- [ ] TODO single place for Ingress domain
- [ ] (Optional) TODO single place for docs link config

TODO single place for all config, maybe a `./config` directory with some symlinks?
