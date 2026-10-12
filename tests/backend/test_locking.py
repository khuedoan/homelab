"""Exercise the generated R2 backend with real clients and a loopback S3 mock."""

import hashlib
import http.server
import json
import os
from pathlib import Path
import shlex
import shutil
import subprocess
import sys
import tempfile
import threading
import time
import unittest
import urllib.parse


ROOT = Path(__file__).resolve().parents[2]


class ObjectStore:
    """Only the S3 operations used by the backend; no authentication emulation."""

    def __init__(self):
        self.objects = {}
        self.events = []
        self.mutex = threading.Lock()

    def snapshot(self):
        with self.mutex:
            return dict(self.objects), list(self.events)

    def handler(self):
        store = self

        class Handler(http.server.BaseHTTPRequestHandler):
            def log_message(self, *_args):
                pass

            def respond(self, status, body=b"", content_type="application/xml"):
                self.send_response(status)
                self.send_header("Content-Type", content_type)
                self.send_header("Content-Length", str(len(body)))
                if status == 200:
                    self.send_header("ETag", '"' + hashlib.md5(body).hexdigest() + '"')
                self.end_headers()
                if self.command != "HEAD":
                    self.wfile.write(body)

            def do_GET(self):
                parsed = urllib.parse.urlsplit(self.path)
                path = urllib.parse.unquote(parsed.path)
                with store.mutex:
                    if "list-type" in urllib.parse.parse_qs(parsed.query):
                        self.respond(200, b'<ListBucketResult xmlns="http://s3.amazonaws.com/doc/2006-03-01/"><IsTruncated>false</IsTruncated><KeyCount>0</KeyCount></ListBucketResult>')
                    elif path in store.objects:
                        self.respond(200, store.objects[path], "application/json")
                    else:
                        self.respond(404, b"<Error><Code>NoSuchKey</Code></Error>")

            do_HEAD = do_GET

            def do_PUT(self):
                path = urllib.parse.unquote(urllib.parse.urlsplit(self.path).path)
                body = self.rfile.read(int(self.headers.get("Content-Length", 0)))
                condition = self.headers.get("If-None-Match")
                with store.mutex:
                    status = 200
                    if path.endswith(".tflock") and condition != "*":
                        status = 400
                    elif condition == "*" and path in store.objects:
                        status = 412
                    store.events.append(("PUT", path, status, condition))
                    if status == 200:
                        store.objects[path] = body
                        self.respond(200)
                    else:
                        self.respond(status, b"<Error><Code>PreconditionFailed</Code></Error>")

            def do_DELETE(self):
                path = urllib.parse.unquote(urllib.parse.urlsplit(self.path).path)
                with store.mutex:
                    store.events.append(("DELETE", path, 204, None))
                    store.objects.pop(path, None)
                    self.respond(204)

        return Handler


class BackendLockingTest(unittest.TestCase):
    def setUp(self):
        for binary in ("tofu", "terragrunt", "git"):
            self.assertIsNotNone(shutil.which(binary), f"{binary} is required; enter nix develop")
        self.temp = tempfile.TemporaryDirectory(prefix="homelab-state-test-")
        self.addCleanup(self.temp.cleanup)
        self.directory = Path(self.temp.name)
        self.env = {
            key: value for key, value in os.environ.items()
            if not key.startswith(("AWS_", "TF_", "TG_", "TERRAGRUNT_", "CLOUDFLARE_"))
        }
        self.env.update({
            "HOME": str(self.directory),
            "AWS_ACCESS_KEY_ID": "mock-access-key",
            "AWS_SECRET_ACCESS_KEY": "mock-secret-key",
            "AWS_EC2_METADATA_DISABLED": "true",
            "AWS_CONFIG_FILE": os.devnull,
            "AWS_SHARED_CREDENTIALS_FILE": os.devnull,
            "TF_IN_AUTOMATION": "true",
            "CHECKPOINT_DISABLE": "1",
            "NO_PROXY": "127.0.0.1,localhost",
            "no_proxy": "127.0.0.1,localhost",
            "TG_LOG_LEVEL": "error",
        })
        self.store = ObjectStore()
        server = http.server.ThreadingHTTPServer(("127.0.0.1", 0), self.store.handler())
        self.addCleanup(server.server_close)
        self.addCleanup(server.shutdown)
        thread = threading.Thread(target=server.serve_forever, daemon=True)
        thread.start()
        self.endpoint = f"http://127.0.0.1:{server.server_port}"
        self.fixture = self.directory / "repo"
        (self.fixture / "infra").mkdir(parents=True)
        shutil.copyfile(ROOT / "infra/root.hcl", self.fixture / "infra/root.hcl")
        self.run_command(["git", "init", "--quiet"], self.fixture)
        bin_directory = self.directory / "bin"
        bin_directory.mkdir()
        token = self.directory / "oauth.json"
        token.write_text(json.dumps({"oauth_token": "mock-cloudflare-token"}))
        identity = json.dumps({
            "authSource": f"OAuth token from {token}",
            "accounts": [{"id": "mock-account"}],
        })
        cf = bin_directory / "cf"
        cf.write_text("#!/bin/sh\nprintf '%s\\n' " + shlex.quote(identity) + "\n")
        cf.chmod(0o700)
        self.env["PATH"] = str(bin_directory) + os.pathsep + self.env["PATH"]

    def run_command(self, args, cwd, expected=0):
        result = subprocess.run(
            args, cwd=cwd, env=self.env, capture_output=True, text=True, timeout=60,
        )
        if expected is not None:
            self.assertEqual(result.returncode, expected, result.stdout + result.stderr)
        return result

    def render_backend(self, environment, unit):
        directory = self.fixture / "infra" / environment / unit
        directory.mkdir(parents=True)
        (directory / "terragrunt.hcl").write_text(
            'include "root" { path = find_in_parent_folders("root.hcl") }\n'
        )
        result = self.run_command(["terragrunt", "render", "--json"], directory)
        rendered = json.loads(result.stdout)
        self.assertEqual(rendered["terraform_binary"], "tofu")
        self.assertNotIn("remote_state", rendered)
        backend = rendered["generate"]["backend"]["contents"]
        self.assertIn('required_version = ">= 1.10.0"', backend)
        self.assertRegex(backend, r"use_lockfile\s*=\s*true")
        self.assertNotIn("mock-cloudflare-token", backend)
        self.assertNotIn("mock-secret-key", backend)
        self.assertNotIn("-reconfigure", json.dumps(rendered["terraform"]))
        # Change only the endpoint in the actual generated configuration.
        self.assertEqual(backend.count("https://mock-account.r2.cloudflarestorage.com"), 1)
        return backend.replace("https://mock-account.r2.cloudflarestorage.com", self.endpoint)

    def client(self, name, backend, hold=False):
        directory = self.directory / name
        directory.mkdir()
        (directory / "backend.tf").write_text(backend)
        provisioner = ""
        if hold:
            script = self.directory / "hold.py"
            script.write_text(
                "import pathlib, time\n"
                f"pathlib.Path({str(self.directory / 'started')!r}).touch()\n"
                "deadline = time.monotonic() + 60\n"
                f"while not pathlib.Path({str(self.directory / 'release')!r}).exists():\n"
                "    if time.monotonic() >= deadline: raise SystemExit('release timed out')\n"
                "    time.sleep(0.05)\n"
            )
            command = shlex.join([sys.executable, str(script)])
            provisioner = 'provisioner "local-exec" { command = ' + json.dumps(command) + ' }'
        (directory / "main.tf").write_text(
            'resource "terraform_data" "test" {\n  input = "fixture"\n' + provisioner + '\n}\n'
        )
        self.run_command(["tofu", "init", "-input=false", "-no-color"], directory)
        return directory

    def test_reuses_existing_object_without_migration(self):
        backend = self.render_backend("staging", "cloudflare")
        key = "/homelab-staging-tfstate/cloudflare/tfstate.json"
        previous = {
            "version": 4,
            "terraform_version": "1.10.0",
            "serial": 12,
            "lineage": "9cf03e68-2167-4c62-97c4-90ef56c36efa",
            "outputs": {"existing": {"value": "preserved", "type": "string"}},
            "resources": [],
        }
        original = json.dumps(previous).encode()
        with self.store.mutex:
            self.store.objects[key] = original
        directory = self.client("existing-state", backend)
        result = self.run_command(["tofu", "state", "pull"], directory)
        current = json.loads(result.stdout)
        for field in ("lineage", "serial", "outputs", "resources"):
            self.assertEqual(current[field], previous[field])
        objects, events = self.store.snapshot()
        self.assertEqual(objects[key], original)
        self.assertFalse(any(method == "PUT" and path == key
                             for method, path, _status, _condition in events))

    def test_lock_contention_release_and_isolation(self):
        backend = self.render_backend("staging", "cloudflare")
        owner = self.client("owner", backend, hold=True)
        contender = self.client("contender", backend)
        other_unit = self.client("other-unit", self.render_backend("staging", "metal"))
        other_environment = self.client("other-environment", self.render_backend("production", "cloudflare"))
        lock_path = "/homelab-staging-tfstate/cloudflare/tfstate.json.tflock"
        process = subprocess.Popen(
            ["tofu", "apply", "-auto-approve", "-input=false", "-no-color"],
            cwd=owner, env=self.env, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True,
        )
        try:
            deadline = time.monotonic() + 30
            while not (self.directory / "started").exists():
                if process.poll() is not None:
                    self.fail(process.communicate()[0])
                if time.monotonic() >= deadline:
                    self.fail("owner did not enter its locked apply")
                time.sleep(0.05)
            objects, _ = self.store.snapshot()
            self.assertIn(lock_path, objects)
            original_lock = objects[lock_path]
            holder_id = json.loads(original_lock)["ID"]

            result = self.run_command(
                ["tofu", "plan", "-input=false", "-lock-timeout=0s", "-no-color"],
                contender, expected=None,
            )
            self.assertNotEqual(result.returncode, 0, result.stdout + result.stderr)
            self.assertIn("Error acquiring the state lock", result.stdout + result.stderr)
            self.assertIn(holder_id, result.stdout + result.stderr)
            self.assertEqual(self.store.snapshot()[0][lock_path], original_lock)

            result = self.run_command(
                ["tofu", "force-unlock", "-force", "not-the-holder-id"], contender, expected=None,
            )
            self.assertNotEqual(result.returncode, 0)
            self.assertEqual(self.store.snapshot()[0][lock_path], original_lock)

            for independent in (other_unit, other_environment):
                self.run_command(["tofu", "plan", "-input=false", "-no-color"], independent)
                self.assertEqual(self.store.snapshot()[0][lock_path], original_lock)
        finally:
            (self.directory / "release").touch()
            try:
                output, _ = process.communicate(timeout=30)
            except subprocess.TimeoutExpired:
                process.kill()
                process.communicate()
                raise
        self.assertEqual(process.returncode, 0, output)
        self.assertNotIn(lock_path, self.store.snapshot()[0])
        self.run_command(["tofu", "plan", "-input=false", "-no-color"], contender)
        objects, events = self.store.snapshot()
        self.assertFalse(any(path.endswith(".tflock") for path in objects))
        state = json.loads(objects[lock_path.removesuffix(".tflock")])
        self.assertEqual(state["resources"][0]["type"], "terraform_data")
        self.assertIn(("PUT", lock_path, 412, "*"), events)
        locks = {path for method, path, status, condition in events
                 if method == "PUT" and path.endswith(".tflock") and status == 200 and condition == "*"}
        self.assertEqual(locks, {
            lock_path,
            "/homelab-staging-tfstate/metal/tfstate.json.tflock",
            "/homelab-production-tfstate/cloudflare/tfstate.json.tflock",
        })


if __name__ == "__main__":
    unittest.main()
