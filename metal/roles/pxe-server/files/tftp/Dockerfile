FROM alpine:3

RUN apk add tftp-hpa gettext

COPY grub.cfg /var/lib/tftpboot/grub.cfg


COPY start.sh /start.sh

CMD [ "/start.sh" ]
