package cluster

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
	coreclient "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

func testSigner(t *testing.T) ssh.Signer {
	t.Helper()
	_, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := ssh.NewSignerFromKey(key)
	if err != nil {
		t.Fatal(err)
	}
	return signer
}

type testRemoteFiles map[string][]byte

func (files testRemoteFiles) Fileread(request *sftp.Request) (io.ReaderAt, error) {
	data, ok := files[request.Filepath]
	if !ok {
		return nil, os.ErrNotExist
	}
	return bytes.NewReader(data), nil
}

type testSSHService struct {
	t        *testing.T
	files    testRemoteFiles
	forwards map[string]string
	workers  sync.WaitGroup
}

func (service *testSSHService) start(work func()) {
	service.workers.Add(1)
	go func() {
		defer service.workers.Done()
		work()
	}()
}

func (service *testSSHService) check(err error) {
	service.t.Helper()
	// Clients intentionally cancel transports and reject host keys in these tests.
	if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, net.ErrClosed) && !errors.Is(err, io.ErrClosedPipe) && !errors.Is(err, syscall.ECONNRESET) {
		service.t.Errorf("SSH fixture protocol failure: %v", err)
	}
}

func (service *testSSHService) serve(socket net.Conn, config *ssh.ServerConfig) {
	conn, channels, requests, err := ssh.NewServerConn(socket, config)
	if err != nil {
		service.check(err)
		closeTransport(socket)
		return
	}
	defer closeTransport(conn)
	go ssh.DiscardRequests(requests)
	for request := range channels {
		switch request.ChannelType() {
		case "direct-tcpip":
			service.start(func() { service.forward(request) })
		case "session":
			service.start(func() { service.session(request) })
		default:
			service.check(request.Reject(ssh.UnknownChannelType, "unsupported"))
		}
	}
}

func (service *testSSHService) forward(request ssh.NewChannel) {
	var target struct {
		Host       string
		Port       uint32
		OriginHost string
		OriginPort uint32
	}
	if err := ssh.Unmarshal(request.ExtraData(), &target); err != nil {
		service.check(request.Reject(ssh.Prohibited, "unexpected target"))
		return
	}
	destination, ok := service.forwards[net.JoinHostPort(target.Host, fmt.Sprint(target.Port))]
	if !ok {
		service.check(request.Reject(ssh.Prohibited, "unexpected target"))
		return
	}
	upstream, err := net.DialTimeout("tcp", destination, time.Second)
	if err != nil {
		service.check(request.Reject(ssh.ConnectionFailed, "unavailable"))
		return
	}
	defer closeTransport(upstream)
	channel, requests, err := request.Accept()
	if err != nil {
		service.check(err)
		return
	}
	defer closeTransport(channel)
	go ssh.DiscardRequests(requests)
	done := make(chan struct{})
	go func() {
		defer close(done)
		_, err := io.Copy(upstream, channel)
		service.check(err)
		closeTransport(upstream)
	}()
	_, err = io.Copy(channel, upstream)
	service.check(err)
	closeTransport(channel)
	<-done
}

func (service *testSSHService) session(request ssh.NewChannel) {
	channel, requests, err := request.Accept()
	if err != nil {
		service.check(err)
		return
	}
	defer closeTransport(channel)
	for request := range requests {
		var subsystem struct{ Name string }
		valid := request.Type == "subsystem" && ssh.Unmarshal(request.Payload, &subsystem) == nil && subsystem.Name == "sftp"
		if err := request.Reply(valid, nil); err != nil {
			service.check(err)
			return
		}
		if !valid {
			continue
		}
		server := sftp.NewRequestServer(channel, sftp.Handlers{FileGet: service.files})
		service.check(server.Serve())
		service.check(server.Close())
		return
	}
}

func testSSHServer(t *testing.T, files testRemoteFiles, forwards map[string]string) (host, sshServers, ssh.PublicKey) {
	t.Helper()
	serverKey, userKey := testSigner(t), testSigner(t)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	service := &testSSHService{t: t, files: files, forwards: forwards}
	accepted := make(chan struct{})
	t.Cleanup(func() {
		if err := listener.Close(); err != nil {
			t.Error(err)
		}
		<-accepted
		service.workers.Wait()
	})
	config := &ssh.ServerConfig{PublicKeyCallback: func(meta ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
		if meta.User() != "root" || !bytes.Equal(key.Marshal(), userKey.PublicKey().Marshal()) {
			return nil, fmt.Errorf("unauthorized")
		}
		return nil, nil
	}}
	config.AddHostKey(serverKey)
	go func() {
		defer close(accepted)
		for {
			socket, err := listener.Accept()
			if err != nil {
				service.check(err)
				return
			}
			service.start(func() { service.serve(socket, config) })
		}
	}()

	node := host{"test", netip.MustParseAddrPort(listener.Addr().String())}
	verify := testHostVerifier(t, node.address.String(), serverKey.PublicKey())
	return node, sshServers{&ssh.ClientConfig{User: "root", Auth: []ssh.AuthMethod{ssh.PublicKeys(userKey)}, HostKeyCallback: verify}}, serverKey.PublicKey()
}

func testHostVerifier(t *testing.T, address string, key ssh.PublicKey) ssh.HostKeyCallback {
	t.Helper()
	known := filepath.Join(t.TempDir(), "known_hosts")
	line := ""
	if key != nil {
		line = knownhosts.Line([]string{address}, key) + "\n"
	}
	if err := os.WriteFile(known, []byte(line), 0600); err != nil {
		t.Fatal(err)
	}
	verify, err := knownhosts.New(known)
	if err != nil {
		t.Fatal(err)
	}
	return verify
}

func TestNativeSSHRequiresTrustedHostKey(t *testing.T) {
	node, remote, serverKey := testSSHServer(t, nil, nil)
	for _, tc := range []struct {
		name    string
		key     ssh.PublicKey
		trusted bool
	}{
		{"trusted", serverKey, true},
		{"changed host key", testSigner(t).PublicKey(), false},
		{"unknown host key", nil, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			remote.config.HostKeyCallback = testHostVerifier(t, node.address.String(), tc.key)
			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			defer cancel()
			conn, err := remote.connect(ctx, node)
			if conn != nil {
				defer conn.close()
			}
			if (err == nil) != tc.trusted {
				t.Fatalf("trusted = %v, connect error = %v", tc.trusted, err)
			}
		})
	}
}

func TestNativeSSHReadsSFTP(t *testing.T) {
	node, remote, _ := testSSHServer(t, testRemoteFiles{"/payload": []byte("native-sftp-payload")}, nil)
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	conn, err := remote.connect(ctx, node)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.close()
	if data, err := readFile(conn.files, "/payload"); err != nil || data != "native-sftp-payload" {
		t.Fatalf("SFTP read: %q, %v", data, err)
	}
}

func testWriteResponse(t *testing.T, w http.ResponseWriter, body string) {
	t.Helper()
	if _, err := io.WriteString(w, body); err != nil {
		t.Errorf("write test API response: %v", err)
	}
}

func TestNativeSSHTunnelsKubernetesAPI(t *testing.T) {
	apiServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/readyz" {
			testWriteResponse(t, w, "ready")
			return
		}
		if r.URL.Path != "/api/v1/namespaces/kube-system" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		testWriteResponse(t, w, `{"apiVersion":"v1","kind":"Namespace","metadata":{"name":"kube-system","uid":"tunneled-cluster"}}`)
	}))
	defer apiServer.Close()
	address := apiServer.Listener.Addr().String()
	node, remote, _ := testSSHServer(t, nil, map[string]string{address: address})
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	conn, err := remote.connect(ctx, node)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.close()
	ca := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: apiServer.Certificate().Raw})
	api, err := coreclient.NewForConfig(&rest.Config{Host: apiServer.URL, Dial: conn.ssh.DialContext, TLSClientConfig: rest.TLSClientConfig{CAData: ca}})
	if err != nil {
		t.Fatal(err)
	}
	if uid, err := apiIdentity(ctx, api); err != nil || uid != "tunneled-cluster" {
		t.Fatalf("Kubernetes API through SSH: %q, %v", uid, err)
	}
}

func TestNativeSSHCancellationClosesTransport(t *testing.T) {
	node, remote, _ := testSSHServer(t, nil, nil)
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	conn, err := remote.connect(ctx, node)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.close()
	cancel()
	closed := make(chan error, 1)
	go func() { closed <- conn.ssh.Wait() }()
	select {
	case <-closed:
	case <-time.After(time.Second):
		t.Fatal("cancellation did not close SSH transport")
	}
}
