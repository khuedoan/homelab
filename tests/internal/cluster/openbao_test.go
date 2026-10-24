package cluster

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	baokubernetes "github.com/openbao/openbao/api/auth/kubernetes/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/rest"
)

const baoProxyPath = "/api/v1/namespaces/openbao/services/http:openbao:8200/proxy"

func TestOpenBaoUsesExplicitTrustAuthAndProxyPath(t *testing.T) {
	t.Setenv("BAO_ADDR", "https://ambient.invalid")
	t.Setenv("BAO_TOKEN", "ambient-secret")
	t.Setenv("BAO_SKIP_VERIFY", "true")
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer kubernetes-credential", r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case baoProxyPath + "/v1/auth/kubernetes/login":
			assert.Empty(t, r.Header.Get("X-Vault-Token"))
			var login struct{ Role, JWT string }
			assert.NoError(t, json.NewDecoder(r.Body).Decode(&login))
			assert.Equal(t, "apps", login.Role)
			assert.Equal(t, "explicit-service-account-token", login.JWT)
			_, err := w.Write([]byte(`{"auth":{"client_token":"scoped-token"}}`))
			assert.NoError(t, err)
		case baoProxyPath + "/v1/secret/data/owner/record":
			assert.Equal(t, "scoped-token", r.Header.Get("X-Vault-Token"))
			_, err := w.Write([]byte(`{"data":{"data":{"value":"visible"},"metadata":{"version":1}}}`))
			assert.NoError(t, err)
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)
	cluster := baoClusterForServer(server)
	client, err := cluster.OpenBao()
	require.NoError(t, err)
	require.Empty(t, client.Token())
	auth, err := baokubernetes.NewKubernetesAuth("apps", baokubernetes.WithServiceAccountToken("explicit-service-account-token"))
	require.NoError(t, err)
	_, err = client.Auth().Login(t.Context(), auth)
	require.NoError(t, err)
	secret, err := client.KVv2("secret").Get(t.Context(), "owner/record")
	require.NoError(t, err)
	require.Equal(t, "visible", secret.Data["value"])

}

func TestOpenBaoRejectsAnUntrustedServerDespiteAmbientSettings(t *testing.T) {
	t.Setenv("BAO_SKIP_VERIFY", "true")
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte(`{"data":{"data":{"value":"visible"},"metadata":{"version":1}}}`))
		assert.NoError(t, err)
	}))
	t.Cleanup(server.Close)
	untrusted := Cluster{config: &rest.Config{Host: server.URL}}
	client, err := untrusted.OpenBao()
	require.NoError(t, err)

	_, err = client.KVv2("secret").Get(t.Context(), "owner/record")

	var untrustedCA x509.UnknownAuthorityError
	require.ErrorAs(t, err, &untrustedCA, "ambient BAO_SKIP_VERIFY must not bypass TLS verification")
}

func TestOpenBaoDoesNotFollowRedirects(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.Header().Set("Location", "/redirected")
		w.WriteHeader(http.StatusTemporaryRedirect)
	}))
	t.Cleanup(server.Close)
	client, err := baoClusterForServer(server).OpenBao()
	require.NoError(t, err)
	client.SetToken("private-token")

	_, err = client.KVv2("secret").Get(t.Context(), "owner/record")

	require.Error(t, err)
	require.NotContains(t, err.Error(), "private-token")
	require.Equal(t, int32(1), requests.Load(), "redirect must not send a second request")
}

func baoClusterForServer(server *httptest.Server) Cluster {
	return Cluster{config: &rest.Config{
		Host:            server.URL,
		BearerToken:     "kubernetes-credential",
		TLSClientConfig: rest.TLSClientConfig{CAData: pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: server.Certificate().Raw})},
	}}
}
