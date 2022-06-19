# MetalLB

## Deployment

Ansible set up the configuration and ArgoCD deploys MetalLB.

## Router config

In order to use Metallb in BGP mode we must configure Pfsense as router to be able of share
BGP route table and route all network to that loadbalancer IPs. We use
[this tutorial](https://www.danmanners.com/posts/pfsense-bgp-kubernetes/)


- install package ffr

- configure `Services->FRR->Global Settings`:

  ```conf
  [general options]
  enable=x
  default_router_id=192.168.192.1
  ```

- `Services->FRR->BGP->BGP`:

  ```conf
  [bgp router options]
  enable=x
  local_as=64512
  ```

- in `Services->FRR->BGP->Neighbors`:

  ```yaml
  - name: 192.168.192.2
    descr: grigri
    remote_as: 64514
  - name: 192.168.192.11
    descr: k8s-amd64-1
    remote_as: 64513
  - name: 192.168.192.21
    descr: k8s-odroid-hc4-1
    remote_as: 64513
  - name: 192.168.192.22
    descr: k8s-odroid-hc4-2
    remote_as: 64513
  - name: 192.168.192.23
    descr: k8s-odroid-hc4-3
    remote_as: 64513
  - name: 192.168.192.31
    descr: k8s-odroid-c4-1
    remote_as: 64513
  - name: 192.168.192.32
    descr: k8s-odroid-c4-2
    remote_as: 64513
  - name: 192.168.192.41
    descr: k8s-rock64-1
    remote_as: 64513
  - name: 192.168.192.42
    descr: k8s-rock64-2
    remote_as: 64513
  - name: 192.168.192.43
    descr: k8s-rock64-3
    remote_as: 64513
  - name: 192.168.192.44
    descr: k8s-rock64-4
    remote_as: 64513
  - name: 192.168.192.45
    descr: k8s-rock64-5
    remote_as: 64513
  - name: 192.168.192.46
    descr: k8s-rock64-6
    remote_as: 64513
  ```

**Important:** to access metallb network from kubernetes subnet you need to add your host to bgp
Some issues could be experimented if not added as `docker push` not working correctly.
