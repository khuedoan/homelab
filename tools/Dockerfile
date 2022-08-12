FROM archlinux

# Sort mirrors by speed
RUN pacman --sync --refresh --noconfirm \
    reflector \
    && reflector \
    --save /etc/pacman.d/mirrorlist \
    --protocol https \
    --latest 20 \
    --sort rate

RUN pacman --sync --refresh --noconfirm \
    ansible \
    ansible-lint \
    diffutils \
    docker \
    docker-compose \
    git \
    go \
    graphviz \
    grc \
    helm \
    k9s \
    kubectl \
    kustomize \
    libisoburn \
    make \
    mdbook \
    neovim \
    openssh \
    p7zip \
    python \
    python-jinja \
    python-kubernetes \
    python-netaddr \
    python-pip \
    python-pre-commit \
    python-rich \
    shellcheck \
    sudo \
    terraform \
    yamllint

ARG K3D_PATH=/usr/local/bin/k3d
RUN curl -L https://github.com/k3d-io/k3d/releases/latest/download/k3d-linux-amd64 -o $K3D_PATH -s \
    && chmod +x $K3D_PATH

# TODO https://github.com/ansible-collections/community.docker/issues/216
RUN pip install docker-compose
