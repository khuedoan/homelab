#!/bin/sh

docker build -t khuedoan/khuedoan.com-nginx nginx/
docker build -t khuedoan/khuedoan.com-index index/
docker build -t khuedoan/khuedoan.com-gitea gitea/
docker build -t khuedoan/khuedoan.com-mariadb mariadb/
