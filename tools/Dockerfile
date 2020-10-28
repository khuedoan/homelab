FROM centos:8

RUN dnf install -y epel-release

RUN dnf install -y \
    ansible \
    python3-devel \
    libffi-devel \
    gcc \
    openssh \
    openssl-devel \
    python3-libselinux \
    python3-pip \
    which \
    wol

RUN pip3 install kolla-ansible python-openstackclient

RUN mkdir /etc/kolla

COPY ./kolla/ansible.cfg /etc/ansible/
COPY ./kolla/cluster.ini /root/cluster.ini
COPY ./docker-entrypoint.sh /docker-entrypoint.sh

ENTRYPOINT [ "/docker-entrypoint.sh" ]

CMD [ "/root/cluster.ini" ]
