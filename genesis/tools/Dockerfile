FROM alpine

RUN echo "http://dl-cdn.alpinelinux.org/alpine/edge/testing" >> /etc/apk/repositories && apk update

RUN apk add \
        docker \
        helm \
        git \
        kubectl \
        make \
        neovim \
        python3 \
        terraform \
        zsh

RUN git clone --depth 1 https://github.com/khuedoan/dotfiles /root \
        && git clone --depth 1 https://github.com/zdharma/zinit.git /root/.zsh/zinit

WORKDIR /root
