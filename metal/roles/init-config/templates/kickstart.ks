#version=DEVEL

# Do not use graphical install
text

# Keyboard layouts
keyboard --xlayouts='us'
# System language
lang en_US.UTF-8

# Partition clearing information
clearpart --all --drives={{ disk }}
# Partitioning
ignoredisk --only-use={{ disk }}
autopart

# Network information
network --bootproto=static --device={{ network_interface }} --ip={{ hostvars[item]['ansible_host'] }} --gateway={{ ansible_default_ipv4.gateway }} --nameserver={{ dns_server }} --netmask={{ ansible_default_ipv4.netmask }} --ipv6=auto --hostname={{ hostvars[item]['inventory_hostname'] }} --activate

# Use network installation
url --url="http://{{ ansible_default_ipv4.address }}/iso/"
# Disable Setup Agent on first boot
firstboot --disable
# Do not configure the X Window System
skipx
# System services
services --enabled="chronyd"
# System timezone
timezone {{ timezone }} --utc

# Create user (locked by default)
user --groups=wheel --name={{ os_username }}
# Add SSH key
sshkey --username=root "{{ ssh_public_key }}"

# SELinux
selinux --disabled

# Firewall
firewall --disabled

%packages
@^server-product-environment
%end

# Enable some services for Kubernetes
services --enable=iscsid

reboot
