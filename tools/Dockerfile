FROM archlinux

# Sort mirrors by speed
RUN pacman --sync --refresh --noconfirm \
        reflector \
    && reflector \
        --save /etc/pacman.d/mirrorlist \
        --protocol https \
        --latest 20 \
        --sort rate

RUN yes | pacman --sync --refresh \
        ansible \
        argocd \
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
        openssh \
        p7zip \
        python \
        python-kubernetes \
        python-netaddr \
        python-pip \
        sudo \
        terraform

# TODO https://github.com/ansible-collections/community.docker/issues/216
RUN pip install docker-compose
