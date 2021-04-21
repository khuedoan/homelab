FROM archlinux

# TODO investigate why --noconfirm is not working
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
        terragrunt \
        zsh

CMD [ "/bin/zsh" ]
