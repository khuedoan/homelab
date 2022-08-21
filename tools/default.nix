{ pkgs ? import (fetchTarball "https://github.com/NixOS/nixpkgs/archive/00e376e3f3c22d991052dfeaf154c42b09deeb29.tar.gz") {} }:

let
  python-packages = pkgs.python3.withPackages (p: with p; [
    jinja2
    kubernetes
    netaddr
    rich
  ]);
in
pkgs.mkShell {
  buildInputs = [
    pkgs.ansible
    pkgs.ansible-lint
    pkgs.bmake
    pkgs.diffutils
    pkgs.docker
    pkgs.docker-compose
    pkgs.git
    pkgs.go
    pkgs.grc
    pkgs.k9s
    pkgs.kube3d
    pkgs.kubectl
    pkgs.kubernetes-helm
    pkgs.kustomize
    pkgs.libisoburn
    pkgs.neovim
    pkgs.openssh
    pkgs.p7zip
    pkgs.pre-commit
    pkgs.shellcheck
    pkgs.terraform
    pkgs.yamllint

    python-packages
  ];
}
