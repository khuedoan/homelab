# Nextcloud

## Migrate Nextcloud

### Nextcloud

#### apps

```bash
POD_NAME=$(KUBECONFIG=/tmp/kubeconfig.yaml kubectl -n nextcloud get pod -l app.kubernetes.io/component=app --no-headers -o custom-columns=":metadata.name")
KUBECONFIG=/tmp/kubeconfig.yaml kubectl -n nextcloud cp config/www/nextcloud/apps/ $POD_NAME:/tmp/

kubectl -n nextcloud exec -it $POD_NAME -- bash
rm -rf /var/www/html/apps
mv /tmp/apps /var/www/html/apps
chown -R www-data:www-data /var/www/html/apps
```

#### data

```bash
POD_NAME=$(KUBECONFIG=/tmp/kubeconfig.yaml kubectl -n nextcloud get pod -l app.kubernetes.io/component=app --no-headers -o custom-columns=":metadata.name")
for DIR in $(find /datasets/nextcloud/ -maxdepth 1 -mindepth 1); do
    echo $DIR;
    KUBECONFIG=/tmp/kubeconfig.yaml kubectl -n nextcloud cp --retries=5 $DIR $POD_NAME:/var/www/html/data/
done
KUBECONFIG=/tmp/kubeconfig.yaml kubectl -n nextcloud cp --retries=5 /datasets/fotos $POD_NAME:/var/www/html/data/pando/files/
kubectl -n nextcloud exec -it $POD_NAME -- bash
chown -R www-data:www-data /var/www/html/data
```

### Postgres

```bash
docker exec -it nextcloud_db_1 bash

PGPASSWORD="$POSTGRES_PASSWORD" pg_dump "$POSTGRES_DB" -h localhost -U "$POSTGRES_USER" -f /tmp/nextcloud-sqlbkp.bak
exit
docker cp nextcloud_db_1:/tmp/nextcloud-sqlbkp.bak /tmp/
KUBECONFIG=/tmp/kubeconfig.yaml kubectl -n nextcloud cp /tmp/nextcloud-sqlbkp.bak nextcloud-postgres-0:/tmp/

kubectl -n nextcloud exec -it nextcloud-postgres-0 -- bash
su - postgres
psql -d template1 -c "DROP DATABASE \"nextcloud\";"
psql -d template1 -c "CREATE DATABASE \"nextcloud\";"
psql -d nextcloud -f /tmp/nextcloud-sqlbkp.bak
```

## occ command

Run with kubectl:

```bash
chsh -s /bin/bash www-data
su -p www-data -c "php /var/www/html/occ"
```