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
        diffutils \
        docker \
        docker-compose \
        git \
        graphviz \
        helm \
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
        python-rich \
        sudo \
        terraform

# TODO better way to install k3d?
RUN curl -L https://github.com/k3d-io/k3d/releases/download/v5.4.1/k3d-linux-amd64 > /usr/local/bin/k3d \
    && chmod +x /usr/local/bin/k3d

# TODO https://github.com/ansible-collections/community.docker/issues/216
RUN pip install docker-compose
