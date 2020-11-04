#!/bin/sh

ip link add type veth
ip link set dev veth0 up
ip link set dev veth1 up
ip link add name uplinkbr type bridge
ip link set uplinkbr up
ip link set veth0 up
ip link set veth0 master uplinkbr
ip link set dev uplinkbr up
