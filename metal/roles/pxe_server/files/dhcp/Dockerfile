FROM alpine:3

RUN apk add dhcp

RUN touch /var/lib/dhcp/dhcpd.leases

CMD [ "dhcpd", "-d", "-f", "-cf", "/etc/dhcp/dhcpd.conf" ]
