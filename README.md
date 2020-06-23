# Khue's Personal Website

## Features

- Home page and blog written in Rust: khuedoan.com
- Gitea: git.khuedoan.com
- Database: MariaDB
- Reverse proxy: NGINX

## Run locally

Add the following lines to the `/etc/hosts` file:

```
127.0.0.1 khuedoan.com
127.0.0.1 git.khuedoan.com
```

Start the services

`$ MYSQL_ROOT_PASSWORD='PASSWORD_HERE' docker-compose up --build`
