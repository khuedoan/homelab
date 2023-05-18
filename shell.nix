# https://status.nixos.org (nixos-22.11)
{ pkgs ? import (fetchTarball "https://github.com/NixOS/nixpkgs/archive/6c591e7adc51.tar.gz") {} }:

let
  python-packages = pkgs.python3.withPackages (p: with p; [
    jinja2
    kubernetes
    netaddr
    rich
  ]);
in
pkgs.mkShell {
  buildInputs = with pkgs; [
    ansible
    ansible-lint
    bmake
    diffutils
    docker
    docker-compose_1 # TODO upgrade to version 2
    git
    go
    gotestsum
    iproute2
    jq
    k9s
    kube3d
    kubectl
    kubernetes-helm
    kustomize
    libisoburn
    neovim
    openssh
    p7zip
    pre-commit
    shellcheck
    terraform
    yamllint

    python-packages
  ];
}
