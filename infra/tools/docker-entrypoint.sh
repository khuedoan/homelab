#!/bin/sh

set -e

kolla-ansible -i $1 bootstrap-servers
kolla-ansible -i $1 prechecks
kolla-ansible -i $1 deploy
kolla-ansible post-deploy
