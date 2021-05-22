FROM archlinux

# TODO (optimize) Use reflector to generate mirrorlist dynamically
COPY ./mirrorlist /etc/pacman.d/mirrorlist

RUN yes | pacman --sync --refresh \
        ansible \
        docker \
        docker-compose \
        kubectl \
        libisoburn \
        make \
        python \
        python-netaddr \
        terraform \
        zsh

CMD [ "/bin/zsh" ]
