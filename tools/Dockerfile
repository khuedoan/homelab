FROM centos:8

RUN dnf install -y epel-release

RUN dnf install -y ansible python3-devel libffi-devel gcc openssl-devel python3-libselinux python3-pip which

RUN pip3 install kolla-ansible

RUN mkdir /etc/kolla

COPY ./kolla/ansible.cfg /etc/ansible/
COPY ./kolla/cluster.ini /root/cluster.ini
COPY ./docker-entrypoint.sh /docker-entrypoint.sh

ENTRYPOINT [ "/docker-entrypoint.sh" ]

CMD [ "/root/cluster.ini" ]
