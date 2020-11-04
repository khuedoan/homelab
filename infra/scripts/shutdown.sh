#!/bin/sh

ssh root@node0.homelab.local 'efibootmgr --bootnext 0006 && poweroff'
ssh root@node1.homelab.local 'efibootmgr --bootnext 0006 && poweroff'
ssh root@node2.homelab.local 'efibootmgr --bootnext 0006 && poweroff'
ssh root@node3.homelab.local 'efibootmgr --bootnext 0006 && poweroff'
