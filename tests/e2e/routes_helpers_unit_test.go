package e2e

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSameHTTPSOriginAllowsPathRedirects(t *testing.T) {
	original := &http.Request{URL: &url.URL{Scheme: "https", Host: "app.invalid", Path: "/"}}
	redirected := &http.Request{URL: &url.URL{Scheme: "https", Host: "app.invalid", Path: "/login"}}
	require.NoError(t, sameHTTPSOrigin(redirected, []*http.Request{original}))
}

func TestSameHTTPSOriginRejectsOriginChanges(t *testing.T) {
	original := &http.Request{URL: &url.URL{Scheme: "https", Host: "app.invalid", Path: "/"}}
	for _, destination := range []url.URL{
		{Scheme: "http", Host: "app.invalid", Path: "/"},
		{Scheme: "https", Host: "other.invalid", Path: "/"},
		{Scheme: "https", Host: "app.invalid:8443", Path: "/"},
	} {
		t.Run(destination.String(), func(t *testing.T) {
			require.ErrorIs(t, sameHTTPSOrigin(&http.Request{URL: &destination}, []*http.Request{original}), http.ErrUseLastResponse)
		})
	}
}
