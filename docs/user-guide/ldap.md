# LDAP

## Deployment

The deployment of the Lightweight Directory Access Protocol (LDAP) application was carried out in k8s using the [Fedora guide for OpenShift] as reference.

## Migration

The migration process involved recreating the backend and suffix, as follows:

```bash
dsconf localhost backend create --suffix dc=grigri,dc=cloud --be-name userroot --create-suffix
```

To import current values, LDIF files were utilized. First, the current LDIF files were exported using the following command:

```bash
ldapsearch -w 'XXXXXXX' -D 'cn=Directory Manager' -H ldaps://ldap.grigri -b dc=grigri,dc=cloud -LLL "(objectclass=*)"
```

Next, Apache Directory Studio was used for importing the exported files.

## Permissions

The permissions for the LDAP deployment were set in accordance with the guidelines provided by [Red Hat's official documentation]. The following commands were executed:

```bash
ldapmodify -w 'XXXXXXX' -D 'cn=Directory Manager' -H ldaps://ldap.grigri -b dc=grigri,dc=cloud -x
dn: dc=grigri,dc=cloud
changetype: modify
add: aci
aci: (targetattr!="userPassword || aci")(version 3.0; acl "Enable anonymous ac
 cess"; allow (read, search, compare) userdn="ldap:///anyone";)

ldapmodify -w 'XXXXXXX' -D 'cn=Directory Manager' -H ldaps://ldap.grigri -b dc=grigri,dc=cloud -x
dn: ou=People,dc=grigri,dc=cloud
changetype: modify
add: aci
aci: (targetattr="userpassword")(version 3.0; acl "read password access"; allo
 w(read) userdn =  "ldap:///uid=*,ou=Read,ou=ServiceUsers,dc=grigri,dc=cloud";
 )
```

## References

[Fedora guide for OpenShift]: https://directory.fedoraproject.org/docs/389ds/howto/howto-deploy-389ds-on-openshift.html
[Red Hat's official documentation]: https://access.redhat.com/documentation/en-us/red_hat_directory_server/11/html/administration_guide/managing_acis
