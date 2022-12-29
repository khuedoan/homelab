# UniFi

## Config DNS

- `ServicesDNS->ResolverGeneral->Settings->Host Overrides`:

  ```yaml
  - host: unifi-controller
    domain: grigri
    ip_address: 192.168.193.2
    description: Deployed on k8s.grigri in his own load balancer
    additional_names_for_this_host:
    - host: unifi
      domain: grigri
      description: Used for automatically adopt devices
  ```

```bash
» dog unifi-controller.grigri
A unifi-controller.grigri. 1h00m00s   192.168.193.2
» dog unifi
A unifi.grigri. 59m50s   192.168.193.2
```

## Adopt devices

- Factory reset:
  - 10 seconds button
  - ssh and run `syswrapper.sh restore-default`
- Connect to the network
- Click adopt from unifi web interface

**Note**: For wireless mesh use same channel for both APs (disable Channel Optimization). If still doesn't work go to legacy UI and set up same Channel and width in `Settings->RADIOS`

## ssh access

Add SSH key to `System->Network Device SSH Authentication->SSH Keys`

```bash
ssh -oPubkeyAcceptedKeyTypes=+ssh-rsa -oHostkeyAlgorithms=+ssh-rsa -oKexAlgorithms=+diffie-hellman-group1-sha1 pando85@{{ hostname }}
```

## Layer 3 methods for UAP adoption

From this [doc][layer_3_adoption], using SSH to adopt devices from the DMZ network.

After SSH into the device:

```bash
set-inform http://unifi-controller.grigri:8080/inform
```

And then you will see the device now show up for adoption.

[layer_3_adoption]: https://help.ui.com/hc/en-us/articles/204909754-UniFi-Layer-3-methods-for-UAP-adoption-and-management
