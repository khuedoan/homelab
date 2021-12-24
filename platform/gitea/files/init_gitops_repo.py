#!/usr/bin/python

import json
import os
import subprocess
import sys

subprocess.check_call([sys.executable, "-m", "pip", "install", "requests"])

import requests

gitea_host = os.getenv('GITEA_HOST', "gitea-http:3000")
gitea_user = os.environ['GITEA_USER']
gitea_pass = os.environ['GITEA_PASSWORD']
seed_repo = "https://github.com/khuedoan/homelab"
org = "ops"
repo = "homelab"
gitea_url = f"http://{gitea_user}:{gitea_pass}@{gitea_host}"

headers = {
    'Content-Type': 'application/json'
}

data_org = json.dumps({
    'username': org
})

data_repo = json.dumps({
    'clone_addr': seed_repo,
    'uid': 1,
    'repo_owner': org,
    'repo_name': repo,
    'mirror': True
})

resp = requests.post(
    url=f"{gitea_url}/api/v1/admin/users/{gitea_user}/orgs",
    headers=headers,
    data=data_org
)

if resp.status_code == 201:
    print(f"Created organization {org}")
elif resp.status_code == 422:
    print(f"Organization already exists")
else:
    print(f"Error creating organization {org} ({resp.status_code})")
    print(resp.content)
    sys.exit(1)

resp = requests.post(
    url=f"{gitea_url}/api/v1/repos/migrate",
    headers=headers,
    data=data_repo
)

if resp.status_code == 201:
    print(f"Created repository {json.loads(str(resp.content, 'utf8'))['html_url']}")
elif resp.status_code == 409:
    print(f"Repository already exists")
else:
    print(f"Error creating git repository ({resp.status_code})")
    print(resp.content)
    sys.exit(1)
