package state

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"

	cloudflare "github.com/cloudflare/cloudflare-go"
)

type r2ActivationFixture struct {
	api           *cloudflare.API
	client        *http.Client
	probes        atomic.Int32
	minted        atomic.Int32
	tokenResponse string
}

type r2TestTransport struct {
	target *url.URL
	base   http.RoundTripper
}

func (transport r2TestTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	clone := request.Clone(request.Context())
	clone.Host = request.URL.Host
	clone.URL.Scheme, clone.URL.Host = transport.target.Scheme, transport.target.Host
	return transport.base.RoundTrip(clone)
}

func newR2ActivationFixture(t *testing.T, handler http.HandlerFunc) *r2ActivationFixture {
	t.Helper()
	fixture := &r2ActivationFixture{tokenResponse: `{"success":true,"result":{"id":"r2-key-id","value":"r2-one-hour-token"}}`}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fixture.probes.Add(1)
		checkR2Probe(t, r)
		handler(w, r)
	}))
	t.Cleanup(server.Close)
	target, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	fixture.client = &http.Client{Transport: r2TestTransport{target: target, base: server.Client().Transport}}
	fixture.api = credentialsTestAPI(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/permission_groups"):
			writeCredentialsResponse(t, w, `{"success":true,"result":[{"id":"object-write","name":"Workers R2 Storage Bucket Item Write"}]}`)
		case r.Method == http.MethodGet:
			writeCredentialsResponse(t, w, `{"success":true,"result":[]}`)
		case r.Method == http.MethodPost:
			fixture.minted.Add(1)
			writeCredentialsResponse(t, w, fixture.tokenResponse)

		default:
			t.Errorf("unexpected Cloudflare request: %s %s", r.Method, r.URL.Path)
		}
	})
	return fixture
}

func checkR2Probe(t *testing.T, r *http.Request) {
	t.Helper()
	if r.Method != http.MethodHead || r.URL.Path != "/tfstate-test" || r.Host != "test-account.r2.cloudflarestorage.com" {
		t.Errorf("unexpected R2 probe target: %s %s host=%s", r.Method, r.URL.Path, r.Host)
	}
	authorization := r.Header.Get("Authorization")
	if r.Header.Get("X-Amz-Content-Sha256") != "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855" {
		t.Error("probe omitted the S3 empty-body hash")
	}
	if !strings.HasPrefix(authorization, "AWS4-HMAC-SHA256 Credential=r2-key-id/") || !strings.Contains(authorization, "/auto/s3/aws4_request,") || !strings.Contains(authorization, "Signature=") {
		t.Error("probe did not use native SigV4 with the minted identity and R2 scope")
	}
	if strings.Contains(authorization, "oauth-test-token") || strings.Contains(authorization, "r2-one-hour-token") {
		t.Error("probe exposed an API token")
	}
}

func TestR2CredentialsWaitForActivation(t *testing.T) {
	var fixture *r2ActivationFixture
	fixture = newR2ActivationFixture(t, func(w http.ResponseWriter, _ *http.Request) {
		switch fixture.probes.Load() {
		case 1:
			w.WriteHeader(http.StatusUnauthorized)
		case 2:
			w.WriteHeader(http.StatusForbidden)
		default:
			w.WriteHeader(http.StatusOK)
		}
	})
	credentials, err := CreateR2Credentials(t.Context(), fixture.api, fixture.client, "test-account", "tfstate-test")
	if err != nil || credentials.AccessKeyID != "r2-key-id" {
		t.Fatalf("credentials failed: %v", err)
	}
	if fixture.probes.Load() != 3 || fixture.minted.Load() != 1 {
		t.Fatalf("unexpected activation counts: probes=%d minted=%d", fixture.probes.Load(), fixture.minted.Load())
	}
}

func TestR2CredentialsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	fixture := newR2ActivationFixture(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		cancel()
	})
	credentials, err := CreateR2Credentials(ctx, fixture.api, fixture.client, "test-account", "tfstate-test")
	assertNoR2Credentials(t, credentials, err)
	if !errors.Is(err, context.Canceled) || fixture.probes.Load() != 1 {
		t.Fatalf("cancellation did not stop activation: %v probes=%d", err, fixture.probes.Load())
	}
}

func TestR2CredentialsPermanentProbeFailures(t *testing.T) {
	for _, status := range []int{http.StatusMovedPermanently, http.StatusNotFound, http.StatusTooManyRequests, http.StatusInternalServerError} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			fixture := newR2ActivationFixture(t, func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Location", "https://unrelated.invalid/leak")
				w.WriteHeader(status)
			})
			credentials, err := CreateR2Credentials(t.Context(), fixture.api, fixture.client, "test-account", "tfstate-test")
			assertNoR2Credentials(t, credentials, err)
			if fixture.probes.Load() != 1 {
				t.Fatalf("permanent failure retried: probes=%d", fixture.probes.Load())
			}
		})
	}
}

type r2FailingTransport struct{ calls atomic.Int32 }

func (transport *r2FailingTransport) RoundTrip(*http.Request) (*http.Response, error) {
	transport.calls.Add(1)
	return nil, errors.New("oauth-test-token r2-one-hour-token")
}

func TestR2CredentialsTransportFailureIsSanitized(t *testing.T) {
	fixture := newR2ActivationFixture(t, func(http.ResponseWriter, *http.Request) { t.Error("unexpected probe") })
	transport := &r2FailingTransport{}
	fixture.client.Transport = transport
	credentials, err := CreateR2Credentials(t.Context(), fixture.api, fixture.client, "test-account", "tfstate-test")
	assertNoR2Credentials(t, credentials, err)
	if transport.calls.Load() != 1 {
		t.Fatalf("transport failure retried: calls=%d", transport.calls.Load())
	}
}

func TestR2CredentialsRejectMalformedGrant(t *testing.T) {
	fixture := newR2ActivationFixture(t, func(http.ResponseWriter, *http.Request) { t.Error("unexpected probe") })
	fixture.tokenResponse = `{"success":true,"result":{"id":"r2-key-id"}}`
	credentials, err := CreateR2Credentials(t.Context(), fixture.api, fixture.client, "test-account", "tfstate-test")
	assertNoR2Credentials(t, credentials, err)
	if fixture.probes.Load() != 0 {
		t.Fatalf("malformed grant triggered probe: probes=%d", fixture.probes.Load())
	}
}

func assertNoR2Credentials(t *testing.T, credentials R2Credentials, err error) {
	t.Helper()
	if err == nil || credentials != (R2Credentials{}) {
		t.Fatal("failed activation returned credentials or no error")
	}
	for _, secret := range []string{"oauth-test-token", "r2-one-hour-token", "353e3a270c481e79adc4c2b4a033e525f32327241dfd0c2bf65831b16ab48413", "AWS4-HMAC-SHA256"} {
		if strings.Contains(err.Error(), secret) {
			t.Fatal("failure exposed private credentials or signed headers")
		}
	}
}
