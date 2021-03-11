FROM alpine

RUN apk add --no-cache \
        curl \
        docker-cli \
        make \
        neovim \
        python3 \
        tree \
        zsh

RUN apk add --no-cache --repository "http://dl-cdn.alpinelinux.org/alpine/edge/testing" \
        helm \
        kubectl \
        terraform

WORKDIR /root/src
