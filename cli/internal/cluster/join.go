package cluster

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"strings"
)

func (remote sshServers) inspect(ctx context.Context, node host, vip netip.Addr, source credentials) (state nodeState, err error) {
	conn, err := remote.connect(ctx, node)
	if err != nil {
		return "", err
	}
	defer conn.close()
	unlock, err := conn.lock()
	if err != nil {
		return "", err
	}
	defer func() { err = errors.Join(err, unlock()) }()
	return conn.inspect(ctx, vip, source)
}

func (remote sshServers) join(ctx context.Context, node host, vip netip.Addr, source credentials) (err error) {
	conn, err := remote.connect(ctx, node)
	if err != nil {
		return err
	}
	defer conn.close()
	unlock, err := conn.lock()
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, unlock()) }()
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
	state.token, state.hasToken, err = conn.enrollmentToken()
	if err != nil {
		return "", err
	}
	state.serverToken, state.hasState, err = conn.existingState()
	if err != nil {
		return "", err
	}
	if state.hasState && state.active {
		if api, err := conn.api(localAPI); err == nil {
			// k3s can be active before its API becomes ready.
			state.clusterID, _ = apiIdentity(ctx, api)
		}
	}
	return state.classify(source)
}

func (conn *connection) enrollmentToken() (string, bool, error) {
	if found, err := exists(conn.files, tokenDir); err != nil {
		return "", false, err
	} else if found {
		if err := securePath(conn.files, tokenDir, 0700, true); err != nil {
			return "", false, err
		}
	}
	found, err := exists(conn.files, tokenPath)
	if err != nil || !found {
		return "", false, err
	}
	if err := securePath(conn.files, tokenPath, 0600, false); err != nil {
		return "", false, err
	}
	token, err := readFile(conn.files, tokenPath)
	return token, true, err
}

func (conn *connection) existingState() (string, bool, error) {
	var serverToken string
	hasState := false
	if found, err := exists(conn.files, "/var/lib/rancher/k3s/server/token"); err != nil {
		return "", false, err
	} else if found {
		token, err := readFile(conn.files, "/var/lib/rancher/k3s/server/token")
		if err != nil {
			return "", false, err
		}
		serverToken = strings.TrimSpace(token)
		hasState = true
	}
	for _, file := range []string{"/var/lib/rancher/k3s/server/db", "/var/lib/rancher/k3s/server/tls", "/var/lib/rancher/k3s/agent", "/etc/rancher/node/password", "/etc/rancher/k3s/k3s.yaml"} {
		found, err := exists(conn.files, file)
		if err != nil {
			return "", false, err
		}
		hasState = hasState || found
	}
	return serverToken, hasState, nil
}
