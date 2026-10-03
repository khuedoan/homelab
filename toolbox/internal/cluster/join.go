package cluster

import (
	"context"
	"fmt"
	"net/netip"
	"strings"
)

func (remote sshServers) inspect(ctx context.Context, node host, vip netip.Addr, source credentials) (nodeState, error) {
	conn, err := remote.connect(ctx, node)
	if err != nil {
		return "", err
	}
	defer conn.close()
	unlock, err := conn.lock()
	if err != nil {
		return "", err
	}
	defer unlock()
	return conn.inspect(ctx, vip, source)
}

func (remote sshServers) join(ctx context.Context, node host, vip netip.Addr, source credentials) error {
	conn, err := remote.connect(ctx, node)
	if err != nil {
		return err
	}
	defer conn.close()
	unlock, err := conn.lock()
	if err != nil {
		return err
	}
	defer unlock()
	state, err := conn.inspect(ctx, vip, source)
	if err != nil {
		return err
	}
	if state == fresh {
		if err := conn.files.MkdirAll(tokenDir); err != nil {
			return err
		}
		if err := conn.files.Chmod(tokenDir, 0700); err != nil {
			return err
		}
		if err := securePath(conn.files, tokenDir, 0700, true); err != nil {
			return err
		}
		if err := publishToken(conn.files, tokenPath, source.token); err != nil {
			return err
		}
	}
	if state == fresh || state == prepared {
		_, err = conn.run("systemctl start --no-block k3s")
	}
	return err
}

func (conn *connection) inspect(ctx context.Context, vip netip.Addr, source credentials) (nodeState, error) {
	unit, err := conn.run("systemctl show k3s --property=ExecStart --value")
	if err != nil {
		return "", err
	}
	if strings.Contains(unit, "--cluster-init") || !strings.Contains(unit, "--server https://"+vip.String()+":6443") || !strings.Contains(unit, "--token-file "+tokenPath) {
		return "", fmt.Errorf("k3s unit is not configured as a joiner for this cluster")
	}
	active, err := conn.run("systemctl show k3s --property=ActiveState --value")
	if err != nil {
		return "", err
	}
	state := snapshot{active: strings.TrimSpace(active) == "active" || strings.TrimSpace(active) == "activating"}
	if found, err := exists(conn.files, tokenDir); err != nil {
		return "", err
	} else if found {
		if err := securePath(conn.files, tokenDir, 0700, true); err != nil {
			return "", err
		}
	}
	state.hasToken, err = exists(conn.files, tokenPath)
	if err != nil {
		return "", err
	}
	if state.hasToken {
		if err := securePath(conn.files, tokenPath, 0600, false); err != nil {
			return "", err
		}
		state.token, err = readFile(conn.files, tokenPath)
		if err != nil {
			return "", err
		}
	}
	if found, err := exists(conn.files, "/var/lib/rancher/k3s/server/token"); err != nil {
		return "", err
	} else if found {
		token, err := readFile(conn.files, "/var/lib/rancher/k3s/server/token")
		if err != nil {
			return "", err
		}
		state.serverToken = strings.TrimSpace(token)
		state.hasState = true
	}
	for _, file := range []string{"/var/lib/rancher/k3s/server/db", "/var/lib/rancher/k3s/server/tls", "/var/lib/rancher/k3s/agent", "/etc/rancher/node/password", "/etc/rancher/k3s/k3s.yaml"} {
		found, err := exists(conn.files, file)
		if err != nil {
			return "", err
		}
		state.hasState = state.hasState || found
	}
	if state.hasState && state.active {
		if api, err := conn.api(localAPI); err == nil {
			state.clusterID, _ = apiIdentity(ctx, api)
		}
	}
	return state.classify(source)
}
