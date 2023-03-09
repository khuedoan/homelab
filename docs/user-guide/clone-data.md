# Clone data between volumes

## rsync

No dependencies:

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: rsync
  namespace: plex
spec:
  template:
    metadata:
      name: rsync
    spec:
      containers:
        - name: rsync
          image: instrumentisto/rsync-ssh
          command:
            - rsync
            - -av
            - --numeric-ids
            - /src/
            - /dest/
          volumeMounts:
            - name: src
              mountPath: "/src/"
            - name: dest
              mountPath: "/dest/"
      volumes:
        - name: src
          persistentVolumeClaim:
            claimName: config-plex-0
        - name: dest
          persistentVolumeClaim:
            claimName: config-plex-0-zfs
      restartPolicy: Never
      nodeSelector:
        name: grigri
```
