# Khue's Home Lab

## Features

### Infrastructure

![LXD](https://img.shields.io/static/v1?logo=LXC&logoColor=white&label=&message=LXD&color=red)
![OpenNebula](https://img.shields.io/static/v1?logo=OpenNebula&logoColor=white&label=&message=OpenNebula&color=red)
![Ansible](https://img.shields.io/static/v1?logo=Ansible&logoColor=white&label=&message=Ansible&color=red)
![Terraform](https://img.shields.io/static/v1?logo=Terraform&logoColor=white&label=&message=Terraform&color=purple)
![Kubernetes](https://img.shields.io/static/v1?logo=Kubernetes&logoColor=white&label=&message=Kubernetes&color=blue)
![Docker](https://img.shields.io/static/v1?logo=Docker&logoColor=white&label=&message=Docker&color=blue)
![Helm](https://img.shields.io/static/v1?logo=Helm&logoColor=white&label=&message=Helm&color=blue)

### Applications

![Gitea](https://img.shields.io/static/v1?logo=Gitea&logoColor=white&label=&message=Gitea&color=green)
![NGINX](https://img.shields.io/static/v1?logo=NGINX&logoColor=white&label=&message=NGINX&color=green)
![MariaDB](https://img.shields.io/static/v1?logo=MariaDB&logoColor=white&label=&message=MariaDB&color=blue)
![PeerTube](https://img.shields.io/static/v1?logo=PeerTube&logoColor=white&label=&message=PeerTube&color=orange)
![Pastebin](https://img.shields.io/static/v1?logo=Pastebin&logoColor=white&label=&message=Pastebin&color=blue)
![Plex](https://img.shields.io/static/v1?logo=Plex&logoColor=white&label=&message=Plex&color=orange)

## Run locally

Add the following lines to the `/etc/hosts` file:

```
127.0.0.1 khuedoan.com
127.0.0.1 git.khuedoan.com
```

Start the services

`$ MYSQL_ROOT_PASSWORD='PASSWORD_HERE' docker-compose up --build`
