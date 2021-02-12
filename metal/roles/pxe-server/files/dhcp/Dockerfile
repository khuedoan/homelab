FROM alpine:3

RUN apk add dhcp gettext

RUN touch /var/lib/dhcp/dhcpd.leases

COPY dhcpd.conf /etc/dhcp/dhcpd.conf

COPY start.sh /start.sh

CMD [ "/start.sh" ]
