# Manual setup

- [Manual setup](#manual-setup)
  - [Network](#network)
    - [Pfsense](#pfsense)
  - [Servers](#servers)
    - [Flash SDs](#flash-sds)
      - [odroid-hc4](#odroid-hc4)
        - [Bootloader Bypass Method](#bootloader-bypass-method)
      - [Naming convention](#naming-convention)
    - [Troubleshooting](#troubleshooting)
      - [Same mac problem](#same-mac-problem)
      - [No python interpreter found](#no-python-interpreter-found)
    - [Setup](#setup)
    - [Stability problems (currently included 400MHz)](#stability-problems-currently-included-400mhz)
  - [Cluster](#cluster)
  - [Ansible installed services](#ansible-installed-services)

## Network

### Pfsense

- Connect to DMZ `192.168.192.0/24`
- Add DHCP server: range(60-99), but fix agent IPs before add to the cluster.
- *only for controller HA* - Create lb for apiserver (used HaProxy: increase client, server and tunnel* timeouts to 86400000)
- Add DNS entry

**tunnel\***: must be added in `backend->advanced settings->backend pass thru` as
`timeout tunnel 86400s`


## Servers

### Flash SDs

- rock64: [image](https://www.armbian.com/rock64/)
- odroid-c4: [image](https://www.armbian.com/odroid-c4/)
- odroid-hc4: [image](https://www.armbian.com/odroid-hc4/)
- amd64: [usb-stick](https://releases.ubuntu.com/20.04/)
Use script from `scripts/prepare_sdcard.sh` to prepare instances and amd64 install manually.

#### odroid-hc4

**Important**: To be able to boot clean Armbian mainline based u-boot / kernel experiences, you need to remove incompatible Petitboot loader that is shipped with the board.

##### Bootloader Bypass Method

This is now the preferred method.  It is easier, and be performed without a display via SSH

> Install an SD Card with a fresh Armbian image
> Flip device upside down
> With a tool, press and hold down the black button.
> Continue holding button and plug in power to device
> Login to console or SSH and perform follow normal setup procedures
> Verify system can access SPI FLASH device and Erase
> Reboot

```bash
odroidhc4:~:# ls -ltr /dev/mtd*
crw------- 1 root root 90, 0 Nov  6 21:38 /dev/mtd0
brw-rw---- 1 root disk 31, 0 Nov  6 21:38 /dev/mtdblock0
crw------- 1 root root 90, 1 Nov  6 21:38 /dev/mtd0ro
odroidhc4:~:# flash_erase /dev/mtd0 0 0
Erasing 4 Kibyte @ fff000 -- 100 % complete
odroidhc4:~:#
```

#### Naming convention

All nodes must be named with prefix `k8s-{hardware_tag}-{numerical_id}`.
For example:

- k8s-rock64-3
- k8s-odroid-c4-1
- k8s-amd64-1

Also, consider their use case and performance profile. For example, for Ceph nodes:

- k8s-sas-ssd-1
- k8s-hot-storage-2


### Troubleshooting

#### Same mac problem

Editing `/boot/ArmbianEnv.txt` didn't work.

`/etc/network/interfaces`:
```
...
auto eth0
iface eth0 inet dhcp
  hwaddress ether b6:09:a4:06:00:8b
```

#### No python interpreter found

```bash
ln -s /usr/bin/python3 /usr/bin/python
```

### Setup

Use playbook `playbooks/install/so.yml` to setup servers.

```bash
make first-boot
```

### Stability problems (currently included 400MHz)

Change uboot to use ddr 333 MHz frequency.

Compile uboot binaries or download from here: https://nextcloud.grigri.cloud/f/451602

```
git clone https://github.com/armbian/build.git
cd build
# edit `config/sources/families/include/rockchip64_common.inc` with this values for rock64:
#      BOOT_USE_BLOBS=yes
#      BOOT_SOC=rk3328
#      DDR_BLOB='rk33/rk3328_ddr_333MHz_v1.16.bin'
#      MINILOADER_BLOB='rk33/rk322xh_miniloader_v2.50.bin'
#      BL31_BLOB='rk33/rk322xh_bl31_v1.44.elf'

./compile.sh docker
# switch to expert (to see rock64) and compile u-boot for rock64
docker cp $CONTAINER_ID:/root/armbian/cache/sources/u-boot/v2020.10/uboot.img .
docker cp $CONTAINER_ID:/root/armbian/cache/sources/u-boot/v2020.10/idbloader.bin .
docker cp $CONTAINER_ID:/root/armbian/cache/sources/u-boot/v2020.10/trust.bin .

```

Then, in rock64 SD card:
```
dd if=idbloader.bin of=/dev/mmcblk0 seek=64 conv=notrunc
dd if=uboot.img of=/dev/mmcblk0 seek=16384 conv=notrunc
dd if=trust.bin of=/dev/mmcblk0 seek=24576 conv=notrunc
sync
```

Ref:
 - [Stability problems](https://forum.armbian.com/topic/15082-rock64-focal-fossa-memory-frequency/?tab=comments#comment-108127)

## Cluster

`playbooks/install/cluster.yml` to setup Kubernetes.

```bash
ansible-playbook playbooks/install/k8s.yml --become
```

## Ansible installed services

1. ceph
2. metallb
