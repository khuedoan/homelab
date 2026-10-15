package e2e

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/khuedoan/homelab/tests/internal/fixture"
	"github.com/stretchr/testify/require"
	authenticationv1 "k8s.io/api/authentication/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func baoRequest(t *testing.T, cluster fixture.Cluster, token, method, path string, input any) (int, []byte) {
	t.Helper()
	request := cluster.Client.CoreV1().RESTClient().Verb(method).
		AbsPath("/api/v1/namespaces/openbao/services/http:openbao:8200/proxy/v1/"+path).
		SetHeader("X-Vault-Token", token)
	if input != nil {
		body, err := json.Marshal(input)
		require.NoError(t, err)
		request = request.SetHeader("Content-Type", "application/json").Body(body)
	}
	status := 0
	body, _ := request.Do(t.Context()).StatusCode(&status).Raw()
	require.NotZero(t, status, "OpenBao request did not receive an HTTP response")
	t.Logf("OpenBao %s %s returned HTTP %d", method, path, status)
	return status, body
}

func baoLogin(t *testing.T, cluster fixture.Cluster, namespace, account string) (int, string) {
	t.Helper()
	jwt, err := cluster.Client.CoreV1().ServiceAccounts(namespace).CreateToken(t.Context(), account,
		&authenticationv1.TokenRequest{Spec: authenticationv1.TokenRequestSpec{
			Audiences: []string{"https://kubernetes.default.svc"}, ExpirationSeconds: ptr.To(int64(600)),
		}}, metav1.CreateOptions{})
	require.NoError(t, err)
	status, body := baoRequest(t, cluster, "", http.MethodPost, "auth/kubernetes/login", map[string]string{
		"role": "apps", "jwt": jwt.Status.Token,
	})
	if status != http.StatusOK {
		return status, ""
	}
	var response struct {
		Auth struct {
			Token string `json:"client_token"`
		} `json:"auth"`
	}
	require.NoError(t, json.Unmarshal(body, &response))
	require.NotEmpty(t, response.Auth.Token)
	return status, response.Auth.Token
}

func checkSecrets(t *testing.T, cluster fixture.Cluster) {
	t.Helper()
	owner := fixture.Namespace(t, cluster.Client)
	other := fixture.Namespace(t, cluster.Client)
	for _, namespace := range []string{owner, other} {
		_, err := cluster.Client.CoreV1().ServiceAccounts(namespace).Create(t.Context(),
			&corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Name: "default"}}, metav1.CreateOptions{})
		if err != nil {
			require.True(t, apierrors.IsAlreadyExists(err), "default ServiceAccount must exist")
		}
	}
	unseal, err := cluster.Client.CoreV1().Secrets("openbao").Get(t.Context(), "openbao-unseal", metav1.GetOptions{})
	require.NoError(t, err)
	root := string(unseal.Data["vault-root"])
	require.NotEmpty(t, root)
	path := owner + "/verification"
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		request := cluster.Client.CoreV1().RESTClient().Delete().
			AbsPath("/api/v1/namespaces/openbao/services/http:openbao:8200/proxy/v1/secret/metadata/"+path).
			SetHeader("X-Vault-Token", root)
		if err := request.Do(ctx).Error(); err != nil {
			t.Errorf("remove OpenBao test record failed")
		}
	})
	status, _ := baoRequest(t, cluster, root, http.MethodPost, "secret/data/"+path,
		map[string]any{"data": map[string]string{"value": "homelab-secret-injection"}})
	require.Equal(t, http.StatusOK, status)
	status, ownToken := baoLogin(t, cluster, owner, "default")
	require.Equal(t, http.StatusOK, status)
	status, body := baoRequest(t, cluster, ownToken, http.MethodGet, "secret/data/"+path, nil)
	require.Equal(t, http.StatusOK, status)
	var record struct {
		Data struct {
			Data map[string]string `json:"data"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(body, &record))
	require.Equal(t, "homelab-secret-injection", record.Data.Data["value"])
	status, _ = baoRequest(t, cluster, ownToken, http.MethodPost, "secret/data/"+path,
		map[string]any{"data": map[string]string{"value": "unauthorized-write"}})
	require.Equal(t, http.StatusForbidden, status, "apps must not overwrite their own secrets")
	status, otherToken := baoLogin(t, cluster, other, "default")
	require.Equal(t, http.StatusOK, status)
	status, _ = baoRequest(t, cluster, otherToken, http.MethodGet, "secret/data/"+path, nil)
	require.Equal(t, http.StatusForbidden, status, "another namespace must not read the owner secret")
	_, err = cluster.Client.CoreV1().ServiceAccounts(owner).Create(t.Context(),
		&corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Name: "untrusted"}}, metav1.CreateOptions{})
	require.NoError(t, err)
	status, _ = baoLogin(t, cluster, owner, "untrusted")
	require.Equal(t, http.StatusForbidden, status, "the shared apps role must accept only the default account")
	injected, err := cluster.Client.CoreV1().Secrets(owner).Create(t.Context(), &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "injection-check", Labels: map[string]string{"homelab.khuedoan.com/bao-secret": "true"}},
		StringData: map[string]string{"value": "bao:secret/data/" + path + "#value"},
	}, metav1.CreateOptions{})
	require.NoError(t, err)
	require.Equal(t, "homelab-secret-injection", string(injected.Data["value"]), "webhook must resolve the reference")
	_, err = cluster.Client.CoreV1().Secrets(other).Create(t.Context(), &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "cross-namespace-check", Labels: map[string]string{"homelab.khuedoan.com/bao-secret": "true"}},
		StringData: map[string]string{"value": "bao:secret/data/" + path + "#value"},
	}, metav1.CreateOptions{})
	require.ErrorContains(t, err, "permission denied", "webhook must not resolve another namespace's secret")
	t.Log("Verified own-namespace read, write denial, cross-namespace denial, non-default account denial and scoped Secret injection")
}
