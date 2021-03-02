FROM alpine:3

RUN apk add dhcp

RUN touch /var/lib/dhcp/dhcpd.leases

COPY dhcpd.conf /etc/dhcp/dhcpd.conf

CMD [ "dhcpd", "-f", "-cf", "/etc/dhcp/dhcpd.conf" ]
