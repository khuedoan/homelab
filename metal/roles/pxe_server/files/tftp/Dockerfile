FROM alpine:3

RUN apk add tftp-hpa

CMD [ "in.tftpd", "--foreground", "--secure", "/var/lib/tftpboot" ]
