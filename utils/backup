#!/usr/bin/env python

import argparse
from kubernetes import client, config
from kubernetes.client.rest import ApiException

config.load_kube_config()

arg_parser = argparse.ArgumentParser()
arg_parser.add_argument("--namespace", required=True)
arg_parser.add_argument("--pvc", required=True)
arg_parser.add_argument("--action", required=True)
args = arg_parser.parse_args()

namespace = args.namespace
pvc = args.pvc
secret = f"{pvc}-backup-repository"


def apply_custom_resource(api, group, version, plural, name, namespace, body):
    try:
        # Check if the resource exists
        api.get_namespaced_custom_object(
            group=group,
            version=version,
            namespace=namespace,
            plural=plural,
            name=name,
        )

        print(f"Patching {body['kind']} {name}")
        api.patch_namespaced_custom_object(
            group=group,
            version=version,
            namespace=namespace,
            plural=plural,
            name=name,
            body=body,
        )
    except ApiException as e:
        if e.status == 404:
            print(f"Creating {body['kind']} {name}")
            api.create_namespaced_custom_object(
                group=group,
                version=version,
                namespace=namespace,
                plural=plural,
                body=body,
            )
        else:
            raise e


apply_custom_resource(
    api=client.CustomObjectsApi(),
    group="external-secrets.io",
    version="v1beta1",
    plural="externalsecrets",
    name=secret,
    namespace=namespace,
    body={
        "apiVersion": "external-secrets.io/v1beta1",
        "kind": "ExternalSecret",
        "metadata": {
            "name": secret,
            "namespace": namespace,
            "annotations": {
                "app.kubernetes.io/managed-by": "scripts/backup",
            },
        },
        "spec": {
            "secretStoreRef": {
                "kind": "ClusterSecretStore",
                "name": "global-secrets",
            },
            "data": [
                {
                    "remoteRef": {
                        "key": "external",
                        "property": "restic-s3-bucket",
                    },
                    "secretKey": "restic_s3_bucket",
                },
                {
                    "remoteRef": {
                        "key": "external",
                        "property": "restic-s3-access-key",
                    },
                    "secretKey": "restic_s3_access_key",
                },
                {
                    "remoteRef": {
                        "key": "external",
                        "property": "restic-s3-secret-key",
                    },
                    "secretKey": "restic_s3_secret_key",
                },
                {
                    "remoteRef": {
                        "key": "external",
                        "property": "restic-password",
                    },
                    "secretKey": "restic_password",
                },
            ],
            "target": {
                "template": {
                    "data": {
                        "RESTIC_REPOSITORY": f"s3:{{{{ .restic_s3_bucket }}}}/{namespace}/{pvc}",
                        "RESTIC_PASSWORD": "{{ .restic_password }}",
                        "AWS_ACCESS_KEY_ID": "{{ .restic_s3_access_key }}",
                        "AWS_SECRET_ACCESS_KEY": "{{ .restic_s3_secret_key }}",
                    }
                }
            },
        },
    },
)

if args.action == "setup":
    apply_custom_resource(
        api=client.CustomObjectsApi(),
        group="volsync.backube",
        version="v1alpha1",
        plural="replicationsources",
        name=pvc,
        namespace=namespace,
        body={
            "apiVersion": "volsync.backube/v1alpha1",
            "kind": "ReplicationSource",
            "metadata": {
                "name": pvc,
                "namespace": namespace,
                "annotations": {
                    "app.kubernetes.io/managed-by": "scripts/backup",
                },
            },
            "spec": {
                "sourcePVC": pvc,
                "trigger": {"schedule": "*/30 * * * *"},
                "restic": {
                    "pruneIntervalDays": 14,
                    "repository": secret,
                    "retain": {
                        "hourly": 6,
                        "daily": 5,
                        "weekly": 4,
                        "monthly": 2,
                        "yearly": 1,
                    },
                    "copyMethod": "Snapshot",
                },
            },
        },
    )
elif args.action == "restore":
    apply_custom_resource(
        api=client.CustomObjectsApi(),
        group="volsync.backube",
        version="v1alpha1",
        plural="replicationdestinations",
        name=pvc,
        namespace=namespace,
        body={
            "apiVersion": "volsync.backube/v1alpha1",
            "kind": "ReplicationDestination",
            "metadata": {
                "name": pvc,
                "namespace": namespace,
                "annotations": {
                    "app.kubernetes.io/managed-by": "scripts/backup",
                },
            },
            "spec": {
                "trigger": {"manual": "restore-once"},
                "restic": {
                    "repository": secret,
                    "destinationPVC": pvc,
                    "copyMethod": "Direct",
                },
            },
        },
    )
else:
    raise ValueError(f"Invalid action: {args.action}")
