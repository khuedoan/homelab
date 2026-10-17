package state

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	cloudflare "github.com/cloudflare/cloudflare-go"
)

func credentialsTestAPI(t *testing.T, handler http.HandlerFunc) *cloudflare.API {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer oauth-test-token" {
			t.Error("request did not use OAuth identity")
		}
		w.Header().Set("Content-Type", "application/json")
		handler(w, r)
	}))
	t.Cleanup(server.Close)
	api, err := cloudflare.NewWithAPIToken("oauth-test-token", cloudflare.HTTPClient(server.Client()), cloudflare.UsingRetryPolicy(0, 0, 0))
	if err != nil {
		t.Fatal(err)
	}
	api.BaseURL = server.URL
	return api
}

func TestCreateR2Credentials(t *testing.T) {
	before := time.Now()
	var requests []string
	api := credentialsTestAPI(t, func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.RequestURI())
		switch {
		case strings.HasSuffix(r.URL.Path, "/permission_groups"):
			writeCredentialsResponse(t, w, `{"success":true,"result":[{"id":"object-write","name":"Workers R2 Storage Bucket Item Write"}]}`)
		case r.Method == http.MethodGet:
			writeCredentialsResponse(t, w, `{"success":true,"result":[]}`)
		case r.Method == http.MethodPost:
			var token cloudflare.APIToken
			if err := json.NewDecoder(r.Body).Decode(&token); err != nil {
				t.Fatal(err)
			}
			checkCredentialGrant(t, token, before)
			writeCredentialsResponse(t, w, `{"success":true,"result":{"id":"r2-key-id","value":"r2-one-hour-token"}}`)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL)
		}
	})
	credentials, err := CreateR2Credentials(t.Context(), api, newR2ActivationFixture(t, func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }).client, "test-account", "tfstate-test")
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(credentials)
	if err != nil {
		t.Fatal(err)
	}
	var result map[string]any
	if err := json.Unmarshal(encoded, &result); err != nil {
		t.Fatal(err)
	}
	if result["Version"] != float64(1) || result["AccessKeyId"] != "r2-key-id" || result["SecretAccessKey"] != "353e3a270c481e79adc4c2b4a033e525f32327241dfd0c2bf65831b16ab48413" || result["Expiration"] == nil || len(result) != 4 {
		t.Fatalf("unexpected credential-process response: %s", encoded)
	}
	if strings.Contains(string(encoded), "r2-one-hour-token") || strings.Contains(string(encoded), "oauth-test-token") {
		t.Fatal("response exposed an API token")
	}
	if !slices.Equal(requests, []string{"GET /accounts/test-account/tokens/permission_groups", "GET /accounts/test-account/tokens?include_expired=true&per_page=50&page=1", "POST /accounts/test-account/tokens"}) {
		t.Fatalf("requests = %v", requests)
	}
}

func TestCredentialsOnlyPruneExpiredTokensForThisBucket(t *testing.T) {
	expired := time.Now().Add(-time.Hour)
	active := time.Now().Add(time.Hour)
	owned := cloudflare.APIToken{
		ID: "expired-owned", Name: "terraform-tfstate-test", ExpiresOn: &expired,
		Policies: []cloudflare.APITokenPolicies{{Effect: "allow", Resources: map[string]interface{}{"com.cloudflare.edge.r2.bucket.test-account_default_tfstate-test": "*"}, PermissionGroups: []cloudflare.APITokenPermissionGroups{{ID: "object-write"}}}},
	}
	tokens := make([]cloudflare.APIToken, 50)
	for i := range tokens {
		tokens[i] = cloudflare.APIToken{ID: fmt.Sprintf("unrelated-%d", i), Name: "some-other-tool", ExpiresOn: &expired}
	}
	tokens[0] = owned
	tokens[1] = owned
	tokens[1].ID, tokens[1].ExpiresOn = "still-active", &active
	tokens[2] = owned
	tokens[2].ID, tokens[2].ExpiresOn = "permanent", nil
	tokens[3] = owned
	tokens[3].ID, tokens[3].Policies = "wrong-policy", nil
	var requests []string
	api := credentialsTestAPI(t, func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.RequestURI())
		switch {
		case strings.HasSuffix(r.URL.Path, "/permission_groups"):
			writeCredentialsResponse(t, w, `{"success":true,"result":[{"id":"object-write","name":"Workers R2 Storage Bucket Item Write"}]}`)
		case r.Method == http.MethodGet && r.URL.Query().Get("page") == "1":
			if err := json.NewEncoder(w).Encode(map[string]any{"success": true, "result": tokens}); err != nil {
				t.Error(err)
			}
		case r.Method == http.MethodGet:
			writeCredentialsResponse(t, w, `{"success":true,"result":[]}`)
		case r.Method == http.MethodDelete && r.URL.Path == "/accounts/test-account/tokens/expired-owned":
			// Another credential process already pruned this expired token.
			w.WriteHeader(http.StatusNotFound)
			writeCredentialsResponse(t, w, `{"success":false,"errors":[{"code":1000,"message":"not found"}]}`)
		case r.Method == http.MethodPost:
			writeCredentialsResponse(t, w, `{"success":true,"result":{"id":"r2-key-id","value":"r2-one-hour-token"}}`)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL)
		}
	})
	if _, err := CreateR2Credentials(t.Context(), api, newR2ActivationFixture(t, func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }).client, "test-account", "tfstate-test"); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(requests, []string{"GET /accounts/test-account/tokens/permission_groups", "GET /accounts/test-account/tokens?include_expired=true&per_page=50&page=1", "GET /accounts/test-account/tokens?include_expired=true&per_page=50&page=2", "DELETE /accounts/test-account/tokens/expired-owned", "POST /accounts/test-account/tokens"}) {
		t.Fatalf("requests = %v", requests)
	}
}

func TestCredentialsMissingPermissionDoesNotMint(t *testing.T) {
	api := credentialsTestAPI(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || !strings.HasSuffix(r.URL.Path, "/permission_groups") {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL)
		}
		writeCredentialsResponse(t, w, `{"success":true,"result":[]}`)
	})
	credentials, err := CreateR2Credentials(t.Context(), api, newR2ActivationFixture(t, func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }).client, "test-account", "tfstate-test")
	if err == nil || credentials.AccessKeyID != "" {
		t.Fatalf("missing permission returned credentials or no error: %+v, %v", credentials, err)
	}
}

func checkCredentialGrant(t *testing.T, token cloudflare.APIToken, earliest time.Time) {
	t.Helper()
	if token.Name != "terraform-tfstate-test" || token.ExpiresOn == nil || token.ExpiresOn.Before(earliest.Add(59*time.Minute)) || token.ExpiresOn.After(time.Now().Add(time.Hour)) {
		t.Fatalf("unexpected token lifetime or name: %+v", token)
	}
	body, err := json.Marshal(token.Policies)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != `[{"effect":"allow","resources":{"com.cloudflare.edge.r2.bucket.test-account_default_tfstate-test":"*"},"permission_groups":[{"id":"object-write"}]}]` {
		t.Fatalf("unexpected token grant: %s", body)
	}
}

func writeCredentialsResponse(t *testing.T, w http.ResponseWriter, body string) {
	t.Helper()
	if _, err := fmt.Fprint(w, body); err != nil {
		t.Error(err)
	}
}
