package state

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	cloudflare "github.com/cloudflare/cloudflare-go"
)

func TestEnsureR2Bucket(t *testing.T) {
	for _, tc := range []struct {
		name     string
		statuses []int
		methods  []string
		wantErr  bool
	}{
		{"existing", []int{200, 200}, []string{"GET", "GET"}, false},
		{"create then rerun", []int{404, 200, 200}, []string{"GET", "POST", "GET"}, false},
		{"unauthorized", []int{401}, []string{"GET"}, true},
		{"forbidden", []int{403}, []string{"GET"}, true},
		{"server failure", []int{500}, []string{"GET"}, true},
		{"competing create", []int{404, 409, 200, 200}, []string{"GET", "POST", "GET", "GET"}, false},
		{"failed create", []int{404, 403, 404}, []string{"GET", "POST", "GET"}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var methods []string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				methods = append(methods, r.Method)
				wantPath := "/accounts/test-account/r2/buckets"
				if r.Method == "GET" {
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
				if len(methods) > len(tc.statuses) {
					t.Error("unexpected extra request")
					w.WriteHeader(500)
					return
				}
				status := tc.statuses[len(methods)-1]
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(status)
				if status == 200 {
					fmt.Fprint(w, `{"success":true,"result":{"name":"tfstate-test"}}`)
				} else {
					fmt.Fprint(w, `{"success":false,"errors":[{"code":10000,"message":"request failed"}]}`)
				}
			}))
			defer server.Close()
			api, err := cloudflare.NewWithAPIToken("test-token", cloudflare.HTTPClient(server.Client()), cloudflare.UsingRetryPolicy(0, 0, 0))
			if err != nil {
				t.Fatal(err)
			}
			api.BaseURL = server.URL
			err = EnsureR2Bucket(context.Background(), api, "test-account", "tfstate-test")
			if (err != nil) != tc.wantErr {
				t.Fatalf("ensure error = %v, want error %v", err, tc.wantErr)
			}
			if !tc.wantErr {
				if err := EnsureR2Bucket(context.Background(), api, "test-account", "tfstate-test"); err != nil {
					t.Fatal(err)
				}
			}
			if !reflect.DeepEqual(methods, tc.methods) {
				t.Fatalf("requests = %v, want %v", methods, tc.methods)
			}
		})
	}
}
