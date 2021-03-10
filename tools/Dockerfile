FROM alpine

RUN echo "http://dl-cdn.alpinelinux.org/alpine/edge/testing" >> /etc/apk/repositories && apk update

RUN apk add \
        curl \
        docker \
        helm \
        git \
        kubectl \
        make \
        neovim \
        python3 \
        terraform \
        tree \
        zsh

RUN git clone --depth 1 https://github.com/zdharma/zinit.git /root/.zsh/zinit \
        && curl https://raw.githubusercontent.com/khuedoan/dotfiles/master/.zshrc > /root/.zshrc \
        && curl https://raw.githubusercontent.com/khuedoan/dotfiles/master/.p10k.zsh > /root/.p10k.zsh \
        && curl https://raw.githubusercontent.com/khuedoan/dotfiles/master/.aliases > /root/.aliases \
        && zsh -i -c 'exit'

WORKDIR /root
