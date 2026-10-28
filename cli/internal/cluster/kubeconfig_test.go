package cluster

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"k8s.io/client-go/tools/clientcmd"
	clientapi "k8s.io/client-go/tools/clientcmd/api"
)

func TestKubeconfigUsesSelectedContextAndVIP(t *testing.T) {
	input := clientapi.Config{
		CurrentContext: "selected",
		Contexts: map[string]*clientapi.Context{
			"other":    {Cluster: "other", AuthInfo: "other"},
			"selected": {Cluster: "seed", AuthInfo: "admin", Namespace: "apps"},
		},
		Clusters: map[string]*clientapi.Cluster{
			"other": {Server: "https://wrong-cluster:6443"},
			"seed":  {Server: localAPI, CertificateAuthorityData: []byte("ca"), ProxyURL: "http://untrusted-proxy", TLSServerName: "other-host"},
		},
		AuthInfos: map[string]*clientapi.AuthInfo{
			"other": {Exec: &clientapi.ExecConfig{Command: "do-not-execute"}},
			"admin": {ClientCertificateData: []byte("certificate"), ClientKeyData: []byte("private-key"), ClientKey: "/unwanted/local/path"},
		},
	}
	data, err := clientcmd.Write(input)
	if err != nil {
		t.Fatal(err)
	}
	output, err := kubeconfigForEndpoint(data, "https://192.0.2.101:6443")
	if err != nil {
		t.Fatal(err)
	}
	if len(output.Contexts) != 1 || len(output.Clusters) != 1 || len(output.AuthInfos) != 1 || output.CurrentContext != "selected" || output.Contexts["selected"].Namespace != "apps" {
		t.Fatalf("did not export the selected context: %#v", output)
	}
	if output.Clusters["seed"].Server != "https://192.0.2.101:6443" || output.Clusters["seed"].ProxyURL != "" || output.Clusters["seed"].TLSServerName != "" || output.AuthInfos["admin"].ClientKey != "" || string(output.AuthInfos["admin"].ClientKeyData) != "private-key" {
		t.Fatal("wrong endpoint, unsafe references, or lost credentials")
	}
	for _, change := range []func(){
		func() { input.CurrentContext = "missing" },
		func() { input.CurrentContext = "selected"; input.Clusters["seed"].InsecureSkipTLSVerify = true },
		func() {
			input.Clusters["seed"].InsecureSkipTLSVerify = false
			input.AuthInfos["admin"].TokenFile = "/secret/local/file"
		},
		func() {
			input.AuthInfos["admin"].TokenFile = ""
			input.AuthInfos["admin"].Exec = &clientapi.ExecConfig{Command: "do-not-execute"}
		},
	} {
		change()
		data, err := clientcmd.Write(input)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := kubeconfigForEndpoint(data, "https://192.0.2.101:6443"); err == nil {
			t.Fatal("accepted unsafe kubeconfig")
		}
	}
	if _, err := kubeconfigForEndpoint([]byte("CANARY_SECRET: [broken"), localAPI); err == nil || strings.Contains(err.Error(), "CANARY_SECRET") {
		t.Fatalf("invalid kubeconfig leaked content: %v", err)
	}
}

func testClientKubeconfig(t *testing.T, server *httptest.Server) ([]byte, []byte) {
	t.Helper()
	certificate := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: server.Certificate().Raw})
	key, err := x509.MarshalPKCS8PrivateKey(server.TLS.Certificates[0].PrivateKey)
	if err != nil {
		t.Fatal(err)
	}
	privateKey := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: key})
	input, err := clientcmd.Write(clientapi.Config{
		CurrentContext: "default",
		Contexts:       map[string]*clientapi.Context{"default": {Cluster: "default", AuthInfo: "default"}},
		Clusters:       map[string]*clientapi.Cluster{"default": {Server: localAPI, CertificateAuthorityData: certificate}},
		AuthInfos:      map[string]*clientapi.AuthInfo{"default": {ClientCertificateData: certificate, ClientKeyData: privateKey}},
	})
	if err != nil {
		t.Fatal(err)
	}
	return input, privateKey
}

type exportScenario struct {
	hostname string
	vipUID   string
	status   int
	retry    bool
}

type exportFixture struct {
	node       host
	remote     sshServers
	endpoint   string
	privateKey []byte
	requests   *atomic.Int32
}

func testKubeconfigExport(t *testing.T, scenario exportScenario) exportFixture {
	t.Helper()
	requests := &atomic.Int32{}
	vip := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/readyz" {
			status := scenario.status
			if requests.Add(1) == 1 && scenario.retry {
				status = http.StatusServiceUnavailable
			}
			w.WriteHeader(status)
			testWriteResponse(t, w, "CANARY_SECRET")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		testWriteResponse(t, w, `{"apiVersion":"v1","kind":"Namespace","metadata":{"name":"kube-system","uid":"`+scenario.vipUID+`"}}`)
	}))
	vip.TLS = &tls.Config{ClientAuth: tls.RequireAnyClientCert}
	vip.StartTLS()
	t.Cleanup(vip.Close)
	local := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/readyz" {
			testWriteResponse(t, w, "ready")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		testWriteResponse(t, w, `{"apiVersion":"v1","kind":"Namespace","metadata":{"name":"kube-system","uid":"intended-cluster"}}`)
	}))
	local.TLS = &tls.Config{Certificates: vip.TLS.Certificates, ClientAuth: tls.RequireAnyClientCert}
	local.StartTLS()
	t.Cleanup(local.Close)
	input, privateKey := testClientKubeconfig(t, vip)
	node, remote, _ := testSSHServer(t, testRemoteFiles{"/etc/hostname": []byte(scenario.hostname), "/etc/rancher/k3s/k3s.yaml": input}, map[string]string{"127.0.0.1:6443": local.Listener.Addr().String()})
	return exportFixture{node, remote, vip.URL, privateKey, requests}
}

func (fixture exportFixture) checkExport(t *testing.T, output []byte) {
	t.Helper()
	cfg, err := clientcmd.Load(output)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Clusters["default"].Server != fixture.endpoint || !bytes.Equal(cfg.AuthInfos["default"].ClientKeyData, fixture.privateKey) {
		t.Fatal("exported SSH/local endpoint or wrong credentials")
	}
}

func TestNativeKubeconfigExportsAuthenticatedVIP(t *testing.T) {
	fixture := testKubeconfigExport(t, exportScenario{"test", "intended-cluster", http.StatusOK, false})
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	output, err := fixture.remote.kubeconfig(ctx, fixture.node, fixture.endpoint)
	if err != nil {
		t.Fatal(err)
	}
	fixture.checkExport(t, output)
}

func TestNativeKubeconfigRejectsWrongIdentityAndUnreadyVIP(t *testing.T) {
	for _, tc := range []struct {
		name string
		exportScenario
	}{
		{"wrong initializer", exportScenario{"other", "intended-cluster", http.StatusOK, false}},
		{"foreign VIP", exportScenario{"test", "other-cluster", http.StatusOK, false}},
		{"unready VIP", exportScenario{"test", "intended-cluster", http.StatusServiceUnavailable, false}},
		{"unauthorized VIP", exportScenario{"test", "intended-cluster", http.StatusUnauthorized, false}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fixture := testKubeconfigExport(t, tc.exportScenario)
			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			defer cancel()
			output, err := fixture.remote.kubeconfig(ctx, fixture.node, fixture.endpoint)
			if err == nil || len(output) != 0 || strings.Contains(err.Error(), "CANARY_SECRET") || strings.Contains(err.Error(), string(fixture.privateKey)) {
				t.Fatalf("accepted invalid export or leaked credentials: %v", err)
			}
			if tc.hostname != "test" && fixture.requests.Load() != 0 {
				t.Fatal("contacted VIP for the wrong initializer")
			}
		})
	}
}

func TestNativeKubeconfigRetriesVIPReadiness(t *testing.T) {
	fixture := testKubeconfigExport(t, exportScenario{"test", "intended-cluster", http.StatusOK, true})
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	var output []byte
	err := wait(ctx, "VIP readiness", func() error {
		var err error
		output, err = fixture.remote.kubeconfig(ctx, fixture.node, fixture.endpoint)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	fixture.checkExport(t, output)
	if fixture.requests.Load() < 2 {
		t.Fatal("did not retry failed readiness")
	}
}
