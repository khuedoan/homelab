package cluster

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/knownhosts"
)

type sshServers struct{ config *ssh.ClientConfig }

type connection struct {
	ssh   *ssh.Client
	files *sftp.Client
	stop  func() bool
}

func newSSHServers(ctx context.Context) (sshServers, func(), error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return sshServers{}, nil, err
	}
	var paths []string
	for _, path := range []string{filepath.Join(home, ".ssh/known_hosts"), "/etc/ssh/ssh_known_hosts"} {
		if _, err := os.Stat(path); err == nil {
			paths = append(paths, path)
		} else if !os.IsNotExist(err) {
			return sshServers{}, nil, err
		}
	}
	if len(paths) == 0 {
		return sshServers{}, nil, fmt.Errorf("no trusted SSH host keys found; verify node keys before enrollment")
	}
	verifyHost, err := knownhosts.New(paths...)
	if err != nil {
		return sshServers{}, nil, err
	}
	var auth []ssh.AuthMethod
	cleanup := func() {}
	if socket := os.Getenv("SSH_AUTH_SOCK"); socket != "" {
		dialer := net.Dialer{Timeout: 10 * time.Second}
		conn, err := dialer.DialContext(ctx, "unix", socket)
		if err != nil {
			return sshServers{}, nil, fmt.Errorf("connect to SSH agent: %w", err)
		}
		stop := context.AfterFunc(ctx, func() { closeTransport(conn) })
		auth = append(auth, ssh.PublicKeysCallback(agent.NewClient(conn).Signers))
		cleanup = func() { stop(); closeTransport(conn) }
	} else {
		key := os.Getenv("SSH_KEY")
		if key == "" {
			key = filepath.Join(home, ".ssh/id_ed25519")
		}
		data, err := os.ReadFile(key)
		if err != nil {
			return sshServers{}, nil, fmt.Errorf("read SSH key: %w; use SSH_AUTH_SOCK or SSH_KEY", err)
		}
		signer, err := ssh.ParsePrivateKey(data)
		if err != nil {
			return sshServers{}, nil, fmt.Errorf("load SSH key: %w; use an SSH agent for encrypted keys", err)
		}
		auth = append(auth, ssh.PublicKeys(signer))
	}
	return sshServers{&ssh.ClientConfig{User: "root", Auth: auth, HostKeyCallback: verifyHost}}, cleanup, nil
}

func (remote sshServers) connect(ctx context.Context, node host) (*connection, error) {
	address := node.address.String()
	dialer := net.Dialer{Timeout: 10 * time.Second}
	socket, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return nil, fmt.Errorf("connect to %s: %w", node.name, err)
	}
	stop := context.AfterFunc(ctx, func() { closeTransport(socket) })
	if err := socket.SetDeadline(time.Now().Add(10 * time.Second)); err != nil {
		stop()
		closeTransport(socket)
		return nil, fmt.Errorf("set SSH handshake deadline on %s: %w", node.name, err)
	}
	clientConn, channels, requests, err := ssh.NewClientConn(socket, address, remote.config)
	if err != nil {
		stop()
		closeTransport(socket)
		return nil, fmt.Errorf("authenticate %s: %w", node.name, err)
	}
	if err := socket.SetDeadline(time.Time{}); err != nil {
		stop()
		closeTransport(clientConn)
		return nil, fmt.Errorf("clear SSH handshake deadline on %s: %w", node.name, err)
	}
	client := ssh.NewClient(clientConn, channels, requests)
	files, err := sftp.NewClient(client)
	if err != nil {
		stop()
		closeTransport(client)
		return nil, fmt.Errorf("open SFTP on %s: %w", node.name, err)
	}
	return &connection{client, files, stop}, nil
}

func (conn *connection) close() {
	conn.stop()
	closeTransport(conn.files)
	closeTransport(conn.ssh)
}

func closeTransport(transport io.Closer) {
	// Teardown also runs after cancellation or peer closure. It cannot undo a completed operation.
	_ = transport.Close()
}

func (conn *connection) run(command string) (string, error) {
	session, err := conn.ssh.NewSession()
	if err != nil {
		return "", err
	}
	defer closeTransport(session)
	session.Stderr = io.Discard
	output, err := session.Output(command)
	if err != nil {
		return "", fmt.Errorf("remote command failed: %w", err)
	}
	return string(output), nil
}

func (conn *connection) lock() (func() error, error) {
	session, err := conn.ssh.NewSession()
	if err != nil {
		return nil, err
	}
	input, err := session.StdinPipe()
	if err != nil {
		closeTransport(session)
		return nil, err
	}
	output, err := session.StdoutPipe()
	if err != nil {
		closeTransport(session)
		return nil, err
	}
	if err := session.Start(`flock -x /run/lock/homelab-k3s-enroll.lock sh -c 'printf L; cat >/dev/null'`); err != nil {
		closeTransport(session)
		return nil, err
	}
	var ack [1]byte
	if _, err := io.ReadFull(output, ack[:]); err != nil || ack[0] != 'L' {
		closeTransport(session)
		return nil, fmt.Errorf("could not acquire node enrollment lock")
	}
	return func() error {
		defer closeTransport(session)
		err := input.Close()
		if err != nil {
			// Closing the channel releases the lock even if sending stdin EOF failed.
			closeTransport(session)
		}
		return errors.Join(err, session.Wait())
	}, nil
}
