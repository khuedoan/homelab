package cluster

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/netip"
	"reflect"
	"strings"
	"testing"
)

type testServers struct {
	credentials credentials
	nodes       map[string]snapshot
	starts      map[string]int
	failStart   string
	rotated     bool
	reads       int
}

func (remote *testServers) checkHost(context.Context, host) error { return nil }

func (remote *testServers) source(context.Context, host, netip.Addr) (credentials, error) {
	remote.reads++
	if remote.rotated && remote.reads > 1 {
		return credentials{"replacement-secret", "different-cluster"}, nil
	}
	return remote.credentials, nil
}

func (remote *testServers) inspect(_ context.Context, node host, _ netip.Addr, source credentials) (nodeState, error) {
	return remote.nodes[node.name].classify(source)
}

func (remote *testServers) join(ctx context.Context, node host, vip netip.Addr, source credentials) error {
	state, err := remote.inspect(ctx, node, vip, source)
	if err != nil || state == member || state == pending {
		return err
	}
	saved := remote.nodes[node.name]
	saved.hasToken = true
	saved.token = source.token
	remote.nodes[node.name] = saved
	if remote.failStart == node.name {
		return fmt.Errorf("service start interrupted")
	}
	saved.active = true
	saved.hasState = true
	saved.clusterID = source.clusterID
	remote.nodes[node.name] = saved
	remote.starts[node.name]++
	return nil
}

func (remote *testServers) ready(_ context.Context, node host, identity string) error {
	if remote.nodes[node.name].clusterID != identity {
		return fmt.Errorf("node has not joined")
	}
	return nil
}

func (remote *testServers) verify(ctx context.Context, cfg config, source credentials) error {
	for _, node := range cfg.joiners {
		if err := remote.ready(ctx, node, source.clusterID); err != nil {
			return err
		}
	}
	return nil
}

func TestEnrollmentPreflightAndInterruptedRetry(t *testing.T) {
	cfg := config{seed: host{name: "seed"}, joiners: []host{{name: "first"}, {name: "last"}}}
	remote := &testServers{
		credentials: credentials{"canary-secret", "cluster-uid"},
		nodes: map[string]snapshot{
			"first": {},
			"last":  {hasState: true},
		},
		starts: make(map[string]int),
	}
	var progress bytes.Buffer
	if err := enroll(context.Background(), cfg, remote, &progress); err == nil || !strings.Contains(err.Error(), "last") {
		t.Fatalf("accepted foreign datastore: %v", err)
	}
	if remote.nodes["first"].hasToken {
		t.Fatal("published credentials before checking all nodes")
	}
	remote.nodes["last"] = snapshot{}
	remote.failStart = "last"
	if err := enroll(context.Background(), cfg, remote, &progress); err == nil {
		t.Fatal("ignored interrupted startup")
	}
	if remote.nodes["last"].token != "canary-secret" || remote.nodes["first"].clusterID != "cluster-uid" {
		t.Fatal("did not preserve published credentials and completed membership")
	}
	remote.failStart = ""
	for range 2 {
		if err := enroll(context.Background(), cfg, remote, &progress); err != nil {
			t.Fatal(err)
		}
	}
	if !reflect.DeepEqual(remote.starts, map[string]int{"first": 1, "last": 1}) {
		t.Fatalf("restarted existing members: %v", remote.starts)
	}
	if strings.Contains(progress.String(), "canary-secret") {
		t.Fatal("leaked credentials in progress output")
	}
}

func TestEnrollmentRejectsChangedInitializer(t *testing.T) {
	remote := &testServers{credentials: credentials{"secret", "cluster"}, nodes: map[string]snapshot{"joiner": {}}, starts: make(map[string]int), rotated: true}
	err := enroll(context.Background(), config{joiners: []host{{name: "joiner"}}}, remote, io.Discard)
	if err == nil || !strings.Contains(err.Error(), "changed") || remote.nodes["joiner"].hasToken {
		t.Fatalf("accepted changed initializer: %v", err)
	}
}

func TestWaitCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := wait(ctx, "readiness", func() error { return fmt.Errorf("unreachable") })
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("did not propagate cancellation: %v", err)
	}
}
