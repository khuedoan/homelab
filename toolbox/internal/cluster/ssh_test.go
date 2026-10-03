package cluster

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"os"
	"path/filepath"
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

func testSSHServer(t *testing.T, files testRemoteFiles, forwards map[string]string) (host, sshServers, ssh.PublicKey) {
	t.Helper()
	serverKey, userKey := testSigner(t), testSigner(t)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { listener.Close() })
	config := &ssh.ServerConfig{PublicKeyCallback: func(meta ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
		if meta.User() != "root" || !bytes.Equal(key.Marshal(), userKey.PublicKey().Marshal()) {
			return nil, fmt.Errorf("unauthorized")
		}
		return nil, nil
	}}
	config.AddHostKey(serverKey)
	go func() {
		for {
			socket, err := listener.Accept()
			if err != nil {
				return
			}
			go func() {
				defer socket.Close()
				conn, channels, requests, err := ssh.NewServerConn(socket, config)
				if err != nil {
					return
				}
				defer conn.Close()
				go ssh.DiscardRequests(requests)
				for request := range channels {
					if request.ChannelType() == "direct-tcpip" {
						var target struct {
							Host       string
							Port       uint32
							OriginHost string
							OriginPort uint32
						}
						if ssh.Unmarshal(request.ExtraData(), &target) != nil {
							request.Reject(ssh.Prohibited, "unexpected target")
							continue
						}
						destination, ok := forwards[net.JoinHostPort(target.Host, fmt.Sprint(target.Port))]
						if !ok {
							request.Reject(ssh.Prohibited, "unexpected target")
							continue
						}
						upstream, err := net.DialTimeout("tcp", destination, time.Second)
						if err != nil {
							request.Reject(ssh.ConnectionFailed, "unavailable")
							continue
						}
						channel, requests, err := request.Accept()
						if err != nil {
							upstream.Close()
							return
						}
						go ssh.DiscardRequests(requests)
						go func() {
							defer upstream.Close()
							defer channel.Close()
							go io.Copy(upstream, channel)
							io.Copy(channel, upstream)
						}()
						continue
					}
					if request.ChannelType() != "session" {
						request.Reject(ssh.UnknownChannelType, "unsupported")
						continue
					}
					channel, requests, err := request.Accept()
					if err != nil {
						return
					}
					go func() {
						defer channel.Close()
						for request := range requests {
							var subsystem struct{ Name string }
							if request.Type != "subsystem" || ssh.Unmarshal(request.Payload, &subsystem) != nil || subsystem.Name != "sftp" {
								request.Reply(false, nil)
								continue
							}
							request.Reply(true, nil)
							server := sftp.NewRequestServer(channel, sftp.Handlers{FileGet: files})
							server.Serve()
							server.Close()
							return
						}
					}()
				}
			}()
		}
	}()

	node := host{"test", netip.MustParseAddrPort(listener.Addr().String())}
	known := filepath.Join(t.TempDir(), "known_hosts")
	if err := os.WriteFile(known, []byte(knownhosts.Line([]string{node.address.String()}, serverKey.PublicKey())+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	verify, err := knownhosts.New(known)
	if err != nil {
		t.Fatal(err)
	}
	return node, sshServers{&ssh.ClientConfig{User: "root", Auth: []ssh.AuthMethod{ssh.PublicKeys(userKey)}, HostKeyCallback: verify}}, serverKey.PublicKey()
}

func TestNativeSSHTrustSFTPAndCancellation(t *testing.T) {
	apiServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/readyz" {
			w.Write([]byte("ready"))
			return
		}
		if r.URL.Path != "/api/v1/namespaces/kube-system" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"apiVersion":"v1","kind":"Namespace","metadata":{"name":"kube-system","uid":"tunneled-cluster"}}`))
	}))
	defer apiServer.Close()
	address := apiServer.Listener.Addr().String()
	node, remote, serverKey := testSSHServer(t, testRemoteFiles{"/payload": []byte("native-sftp-payload")}, map[string]string{address: address})
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
			known := filepath.Join(t.TempDir(), "known_hosts")
			line := ""
			if tc.key != nil {
				line = knownhosts.Line([]string{node.address.String()}, tc.key) + "\n"
			}
			if err := os.WriteFile(known, []byte(line), 0600); err != nil {
				t.Fatal(err)
			}
			verify, err := knownhosts.New(known)
			if err != nil {
				t.Fatal(err)
			}
			remote.config.HostKeyCallback = verify
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			conn, err := remote.connect(ctx, node)
			if !tc.trusted {
				if err == nil {
					conn.close()
					t.Fatal("accepted untrusted host")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			defer conn.close()
			if data, err := readFile(conn.files, "/payload"); err != nil || data != "native-sftp-payload" {
				t.Fatalf("SFTP read: %q, %v", data, err)
			}
			ca := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: apiServer.Certificate().Raw})
			api, err := coreclient.NewForConfig(&rest.Config{Host: apiServer.URL, Dial: conn.ssh.DialContext, TLSClientConfig: rest.TLSClientConfig{CAData: ca}})
			if err != nil {
				t.Fatal(err)
			}
			if uid, err := apiIdentity(ctx, api); err != nil || uid != "tunneled-cluster" {
				t.Fatalf("Kubernetes API through SSH: %q, %v", uid, err)
			}
			cancel()
			closed := make(chan error, 1)
			go func() { closed <- conn.ssh.Wait() }()
			select {
			case <-closed:
			case <-time.After(time.Second):
				t.Fatal("cancellation did not close SSH transport")
			}
		})
	}
}
