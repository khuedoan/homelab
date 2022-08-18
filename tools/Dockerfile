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
    python-kubernetes \
    python-pip \
    python-pre-commit \
    python-rich \
    shellcheck \
    sudo \
    terraform \
    yamllint

RUN curl -L https://raw.githubusercontent.com/pando85/homelab/master/metal/requirements.txt \
    > /tmp/requirements.txt \
    && pip install -r /tmp/requirements.txt \
    && rm /tmp/requirements.txt

RUN curl -L https://github.com/k3d-io/k3d/releases/download/v5.4.1/k3d-linux-amd64 > /usr/local/bin/k3d \
    && chmod +x /usr/local/bin/k3d
