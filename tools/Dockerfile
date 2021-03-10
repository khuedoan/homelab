FROM alpine

RUN echo "http://dl-cdn.alpinelinux.org/alpine/edge/testing" >> /etc/apk/repositories && apk update

RUN apk add \
        curl \
        docker-cli \
        helm \
        kubectl \
        make \
        neovim \
        python3 \
        terraform \
        tree \
        zsh

WORKDIR /root/src
