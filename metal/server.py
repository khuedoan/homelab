from http.server import BaseHTTPRequestHandler, HTTPServer
import urllib.parse
import subprocess
import threading
import sys
import signal

# TODO clean this shit up, split this to a JSON or some shit
HOST_MAP = {
    "bc:24:11:d0:28:34": ".#metal1",
    "bc:24:11:0d:2f:20": ".#metal2",
}

installed = set()
processes = {}
lock = threading.Lock()
server = None  # HTTPServer instance
pxe_process = None  # PXE server process

class Handler(BaseHTTPRequestHandler):
    def do_POST(self):
        length = int(self.headers['Content-Length'])
        body = self.rfile.read(length).decode()
        data = urllib.parse.parse_qs(body)

        mac = data.get("mac", [""])[0]
        ip = data.get("ip", [""])[0]

        print(f"Got phone-home from {mac} at {ip}")

        if mac in HOST_MAP:
            with lock:
                if mac not in installed and mac not in processes:
                    flake = HOST_MAP[mac]
                    print(f"→ Starting nixos-anywhere for {flake} on {ip}")
                    command = [
                        "nixos-anywhere",
                        "--flake", flake,
                        "--target-host", f"root@{ip}"
                    ]
                    print("Command:", command)

                    proc = subprocess.Popen(command)
                    processes[mac] = proc

                    threading.Thread(target=monitor_install, args=(mac,)).start()
                else:
                    print(f"{mac} already installed or in progress, ignoring")
        else:
            print("Unknown MAC, ignoring")

        self.send_response(200)
        self.end_headers()


def monitor_install(mac):
    proc = processes[mac]
    proc.wait()

    with lock:
        installed.add(mac)
        del processes[mac]

        print(f"Installed servers: {installed}")
        print(f"Pending servers: {set(HOST_MAP.keys()) - installed}")

        if not (set(HOST_MAP.keys()) - installed):
            print("All servers installed, shutting down.")
            stop_server()


def stop_server():
    global server, pxe_process
    if server:
        server.shutdown()
        print("HTTP server shutdown complete.")
    if pxe_process:
        print("Stopping PXE server...")
        pxe_process.send_signal(signal.SIGINT)
        pxe_process.wait()
        print("PXE server stopped.")
    sys.exit(0)


if __name__ == "__main__":
    print("Servers to install:", HOST_MAP)

    # Start PXE server with interactive sudo
    print("Starting PXE server (sudo will prompt)...")
    pxe_process = subprocess.Popen(
        ["sudo", "nix", "run", ".#nixosPxeServer"],
        stdin=sys.stdin, stdout=sys.stdout, stderr=sys.stderr
    )

    # Start HTTP server
    server = HTTPServer(("0.0.0.0", 5000), Handler)
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        stop_server()
