package cluster

import (
	"net/http"
	"strings"
	"time"

	bao "github.com/openbao/openbao/api/v2"
	"k8s.io/client-go/rest"
)

// OpenBao creates a tokenless SDK client through the authenticated Kubernetes service proxy.
func (cluster Cluster) OpenBao() (*bao.Client, error) {
	config := bao.NewConfig()
	config.Address = strings.TrimRight(cluster.config.Host, "/") + "/api/v1/namespaces/openbao/services/http:openbao:8200/proxy"
	config.Timeout = 10 * time.Second
	config.MaxRetries = 0
	config.DisableRedirects = true
	client, err := rest.HTTPClientFor(cluster.RESTConfig())
	if err != nil {
		return nil, err
	}
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	config.HttpClient = client
	return bao.NewClient(config)
}
