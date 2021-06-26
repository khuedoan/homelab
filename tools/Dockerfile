FROM archlinux

# TODO (optimize) Use reflector to generate mirrorlist dynamically
COPY ./mirrorlist /etc/pacman.d/mirrorlist

RUN yes | pacman --sync --refresh \
        ansible \
        ansible-lint \
        diffutils \
        docker \
        docker-compose \
        git \
        kubectl \
        make \
        p7zip \
        python \
        python-netaddr \
        terraform \
        zsh

CMD [ "/bin/zsh" ]
