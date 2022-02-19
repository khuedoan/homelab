# Run commands on multiple nodes

Use [ansible-console](https://docs.ansible.com/ansible/latest/cli/ansible-console.html):

```
ansible-console -i metal/inventories/prod.yml -u root
```

For example:

`root@all (4)[f:5]$ uptime`

```
metal0 | CHANGED | rc=0 >>
 10:52:02 up 2 min,  1 user,  load average: 0.17, 0.15, 0.06
metal1 | CHANGED | rc=0 >>
 10:52:02 up 2 min,  1 user,  load average: 0.14, 0.11, 0.04
metal3 | CHANGED | rc=0 >>
 10:52:02 up 2 min,  1 user,  load average: 0.03, 0.02, 0.00
metal2 | CHANGED | rc=0 >>
 10:52:02 up 2 min,  1 user,  load average: 0.06, 0.06, 0.02
```
