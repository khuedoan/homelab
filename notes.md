# Delete K3s

```shell
systemctl disable k3s.service
systemctl reset-failed k3s.service
systemctl daemon-reload
sudo rm -rf /var/lib/rancher
sudo rm -rf /var/lib/longhorn
sudo rm -rf /etc/systemd/system/k3s.service
sudo rm -rf /etc/rancher
sudo rm -rf /usr/local/bin/k3s
reboot
```
