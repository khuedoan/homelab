FROM archlinux

# TODO (optimize) Use reflector to generate mirrorlist dynamically
COPY ./mirrorlist /etc/pacman.d/mirrorlist

RUN yes | pacman --sync --refresh \
        ansible \
        curl \
        docker \
        git \
        helm \
        kubectl \
        lxd \
        make \
        neovim \
        openssh \
        python \
        python-netaddr \
        terraform \
        zsh

CMD [ "/bin/zsh" ]
