FROM alpine

RUN apk add --no-cache \
        ansible \
        curl \
        docker-cli \
        git \
        make \
        neovim \
        openssh \
        python3 \
        tree \
        zsh

RUN apk add --no-cache --repository "http://dl-cdn.alpinelinux.org/alpine/edge/testing" \
        helm \
        kubectl \
        lxd \
        terraform

ENTRYPOINT [ "/bin/zsh", "-c" ]
