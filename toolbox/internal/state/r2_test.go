package state

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	cloudflare "github.com/cloudflare/cloudflare-go"
)

type r2Response struct {
	method string
	status int
}

func r2TestAPI(t *testing.T, responses ...r2Response) *cloudflare.API {
	t.Helper()
	var mu sync.Mutex
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		requests++
		if requests > len(responses) {
			t.Error("unexpected extra request")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		response := responses[requests-1]
		if r.Method != response.method {
			t.Errorf("request %d method = %s, want %s", requests, r.Method, response.method)
		}
		checkR2Request(t, r)
		if response.status == 0 {
			conn, _, err := w.(http.Hijacker).Hijack()
			if err != nil {
				t.Error(err)
				return
			}
			if err := conn.Close(); err != nil {
				t.Error(err)
			}
			return
		}
		body := `{"success":false,"errors":[{"code":10000,"message":"request failed"}]}`
		if response.status == http.StatusOK {
			body = `{"success":true,"result":{"name":"tfstate-test"}}`
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(response.status)
		if _, err := fmt.Fprint(w, body); err != nil {
			t.Errorf("write R2 response: %v", err)
		}
	}))
	t.Cleanup(func() {
		server.Close()
		mu.Lock()
		defer mu.Unlock()
		if requests != len(responses) {
			t.Errorf("requests = %d, want %d", requests, len(responses))
		}
	})
	api, err := cloudflare.NewWithAPIToken("test-token", cloudflare.HTTPClient(server.Client()), cloudflare.UsingRetryPolicy(0, 0, 0))
	if err != nil {
		t.Fatal(err)
	}
	api.BaseURL = server.URL
	return api
}

func checkR2Request(t *testing.T, r *http.Request) {
	t.Helper()
	wantPath := "/accounts/test-account/r2/buckets"
	if r.Method == http.MethodGet {
		wantPath += "/tfstate-test"
	} else {
		var body struct{ Name string }
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name != "tfstate-test" {
			t.Errorf("create body = %v, %v", body, err)
		}
	}
	if r.URL.Path != wantPath || r.Header.Get("Authorization") != "Bearer test-token" {
		t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
	}
}

func TestEnsureR2BucketExisting(t *testing.T) {
	api := r2TestAPI(t, r2Response{"GET", 200}, r2Response{"GET", 200})
	for range 2 {
		if err := EnsureR2Bucket(t.Context(), api, "test-account", "tfstate-test"); err != nil {
			t.Fatal(err)
		}
	}
}

func TestEnsureR2BucketCreateThenRerun(t *testing.T) {
	api := r2TestAPI(t, r2Response{"GET", 404}, r2Response{"POST", 200}, r2Response{"GET", 200})
	for range 2 {
		if err := EnsureR2Bucket(t.Context(), api, "test-account", "tfstate-test"); err != nil {
			t.Fatal(err)
		}
	}
}

func TestEnsureR2BucketInspectionFailure(t *testing.T) {
	for _, status := range []int{401, 403, 500} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			api := r2TestAPI(t, r2Response{"GET", status})
			err := EnsureR2Bucket(t.Context(), api, "test-account", "tfstate-test")
			if err == nil || !strings.Contains(err.Error(), `inspect state bucket "tfstate-test"`) {
				t.Fatalf("inspection error = %v", err)
			}
			if strings.Contains(err.Error(), "test-token") {
				t.Fatalf("inspection error leaked credentials: %v", err)
			}
		})
	}
}

func TestEnsureR2BucketRechecksFailedCreate(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
	}{
		{"competing create", 409},
		{"lost response", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			api := r2TestAPI(t, r2Response{"GET", 404}, r2Response{"POST", tc.status}, r2Response{"GET", 200}, r2Response{"GET", 200})
			for range 2 {
				if err := EnsureR2Bucket(t.Context(), api, "test-account", "tfstate-test"); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

func TestEnsureR2BucketFailedCreateStillMissing(t *testing.T) {
	api := r2TestAPI(t, r2Response{"GET", 404}, r2Response{"POST", 403}, r2Response{"GET", 404})
	err := EnsureR2Bucket(t.Context(), api, "test-account", "tfstate-test")
	if err == nil || !strings.Contains(err.Error(), `create state bucket "tfstate-test"`) {
		t.Fatalf("creation error = %v", err)
	}
	if strings.Contains(err.Error(), "test-token") {
		t.Fatalf("creation error leaked credentials: %v", err)
	}
}
