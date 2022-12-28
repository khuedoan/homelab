# UniFi

## Config DNS

```bash
» dog unifi-controller.grigri
A unifi-controller.grigri. 1h00m00s   192.168.193.2
```

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
