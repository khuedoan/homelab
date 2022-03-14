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

# TODO https://github.com/ansible-collections/community.docker/issues/216
RUN pip install docker-compose
