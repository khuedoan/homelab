import os
import sys

from cloudflare import APIError, Cloudflare, NotFoundError

account = os.environ["CLOUDFLARE_ACCOUNT_ID"]
bucket = f"homelab-{sys.argv[1]}-tfstate"

try:
    with Cloudflare() as client:
        try:
            client.r2.buckets.get(bucket, account_id=account)
        except NotFoundError:
            client.r2.buckets.create(account_id=account, name=bucket)
except APIError as error:
    raise SystemExit(
        f"State bucket setup failed ({type(error).__name__}). "
        "Check Cloudflare credentials and permissions."
    ) from None

print(f"State bucket ready: {bucket}")
