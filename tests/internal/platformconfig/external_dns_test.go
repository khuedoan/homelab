package platformconfig

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExternalDNSAnnotations(t *testing.T) {
	files := map[string]string{
		"infra/_modules/platform/external_dns.tf": "--annotation-filter=external-dns.kubernetes.io/exclude notin (true)",
		"infra/_modules/platform/istio.tf":        `"external-dns.kubernetes.io/target" = "${var.resource_prefix}-tunnel.${var.domain}"`,
		"apps/blog.yaml":                          "external-dns.kubernetes.io/cloudflare-proxied: 'true'",
		"apps/excalidraw.yaml":                    "external-dns.kubernetes.io/cloudflare-proxied: 'true'",
	}
	for name, want := range files {
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join("../../..", name))
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(data), want) {
				t.Fatalf("missing current ExternalDNS annotation: %s", want)
			}
			if strings.Contains(string(data), "external-dns.alpha.kubernetes.io/") {
				t.Fatal("legacy annotations are disabled by ExternalDNS v0.23")
			}
		})
	}
}
