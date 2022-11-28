#!/bin/bash
set -ex
IMAGE=$1
K8S_HOSTNAME=$2

if [[ $EUID -ne 0 ]]; then
    echo "This script must be run as root"
    exit 1
fi

if [[ -z $IMAGE ]] || [[ -z $K8S_HOSTNAME ]]; then
    echo "Usage: sudo ./prepare_sdcard.sh distros/Armbian_20.05.4_Odroidc4_focal_current_5.6.18.img k8s-1 "
fi

bash -c "dd if=$IMAGE  of=/dev/mmcblk0 bs=1M conv=sync"
sleep 2

if echo "$IMAGE" | grep -i rock64; then
    HOSTNAME=rock64
else
    HOSTNAME=odroid-c4
fi

mount /dev/mmcblk0p1 /mnt
sed -i "s/$HOSTNAME/$K8S_HOSTNAME/g" /mnt/etc/hostname
sed -i "s/$HOSTNAME/$K8S_HOSTNAME/g" /mnt/etc/hosts
rm /mnt/root/.not_logged_in_yet || true
rm /mnt/etc/profile.d/armbian-check-first-login.sh || true
rm /mnt/etc/profile.d/armbian-check-first-login-reboot.sh || true
umount /mnt
