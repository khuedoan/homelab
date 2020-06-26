FROM alpine:3.12.0

ADD https://github.com/w4/bin/releases/download/v1.0.3/bin-v1.0.3-x86_64-unknown-linux-musl.tar.gz /tmp/

RUN tar -xvzf /tmp/bin-*.tar.gz -C /usr/local/bin/ && rm /tmp/bin-*.tar.gz

CMD [ "bin" ]
