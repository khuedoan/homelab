FROM alpine:3

RUN apk add tftp-hpa

COPY ./tftpboot /var/lib/tftpboot

CMD [ "in.tftpd", "--foreground", "--secure", "/var/lib/tftpboot" ]
