# Wallabag

## Fix user login fail

```
cd /var/www/wallabag
su -c "php bin/console cache:clear --env=prod" -s /bin/sh nobody
su -c "php bin/console doctrine:migrations:migrate --no-interaction --env=prod" -s /bin/sh nobody
```
