# Khue's Home Lab

## Table of contents

<!-- vim-markdown-toc GFM -->

* [Features](#features)
    * [Infrastructure](#infrastructure)
    * [Applications](#applications)
* [Run locally](#run-locally)
* [Notes](#notes)

<!-- vim-markdown-toc -->

## Features

### Infrastructure

![Ansible](https://img.shields.io/static/v1?logo=Ansible&logoColor=white&label=&message=Ansible&color=EE0000)
![Docker](https://img.shields.io/static/v1?logo=Docker&logoColor=white&label=&message=Docker&color=2496ED)
![Helm](https://img.shields.io/static/v1?logo=Helm&logoColor=white&label=&message=Helm&color=277A9F)
![Kubernetes](https://img.shields.io/static/v1?logo=Kubernetes&logoColor=white&label=&message=Kubernetes&color=326CE5)
![OpenStack](https://img.shields.io/static/v1?logo=OpenStack&logoColor=white&label=&message=OpenStack&color=ED1944)
![Terraform](https://img.shields.io/static/v1?logo=Terraform&logoColor=white&label=&message=Terraform&color=623CE4)

### Applications

![Gitea](https://img.shields.io/static/v1?logo=Gitea&logoColor=white&label=&message=Gitea&color=609926)
![MariaDB](https://img.shields.io/static/v1?logo=MariaDB&logoColor=white&label=&message=MariaDB&color=003545)
![NGINX](https://img.shields.io/static/v1?logo=NGINX&logoColor=white&label=&message=NGINX&color=269539)
![Pastebin](https://img.shields.io/static/v1?logo=Pastebin&logoColor=white&label=&message=Pastebin&color=02456C)
![PeerTube](https://img.shields.io/static/v1?logo=PeerTube&logoColor=white&label=&message=PeerTube&color=F1680D)
![Plex](https://img.shields.io/static/v1?logo=Plex&logoColor=white&label=&message=Plex&color=E5A00D)

## Run locally

Add the following lines to the `/etc/hosts` file:

```
127.0.0.1 khuedoan.com
127.0.0.1 git.khuedoan.com
```

Start the services

`$ MYSQL_ROOT_PASSWORD='PASSWORD_HERE' docker-compose up --build`

## Notes

- Node hostname must be different and has proper domain (for example `node0.homelab.local`)
