FROM centos:8

RUN dnf install -y epel-release

RUN dnf install -y ansible python3-devel libffi-devel gcc openssl-devel python3-libselinux python3-pip

RUN pip3 install kolla-ansible

RUN mkdir /etc/kolla

COPY ./kolla/cluster.ini /root/cluster.ini

CMD [ "sleep", "10000" ]
