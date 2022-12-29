# https://status.nixos.org (nixpkgs-unstable)
{ pkgs ? import (fetchTarball "https://github.com/NixOS/nixpkgs/archive/e58a7747db96c23b8a977e7c1bbfc5753b81b6fa.tar.gz") {} }:

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
