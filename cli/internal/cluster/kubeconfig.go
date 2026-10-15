package cluster

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	coreclient "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientapi "k8s.io/client-go/tools/clientcmd/api"
)

// Kubeconfig fetches credentials and verifies the VIP from the operator's network.
func Kubeconfig(ctx context.Context, environment string) ([]byte, error) {
	cfg, err := loadConfig(environment)
	if err != nil {
		return nil, err
	}
	remote, close, err := newSSHServers(ctx)
	if err != nil {
		return nil, err
	}
	defer close()
	var data []byte
	err = wait(ctx, "authenticated VIP readiness", func() error {
		var err error
		data, err = remote.kubeconfig(ctx, cfg.seed, "https://"+cfg.vip.String()+":6443")
		return err
	})
	return data, err
}

func (remote sshServers) kubeconfig(ctx context.Context, seed host, endpoint string) ([]byte, error) {
	conn, err := remote.connect(ctx, seed)
	if err != nil {
		return nil, err
	}
	defer conn.close()
	if err := conn.checkHost(seed); err != nil {
		return nil, err
	}
	data, err := readFile(conn.files, "/etc/rancher/k3s/k3s.yaml")
	if err != nil {
		return nil, err
	}
	cfg, err := kubeconfigForEndpoint([]byte(data), endpoint)
	if err != nil {
		return nil, err
	}
	options, err := apiConfig(cfg)
	if err != nil {
		return nil, err
	}
	api, err := coreclient.NewForConfig(options)
	if err != nil {
		return nil, fmt.Errorf("invalid kubeconfig credentials")
	}
	identity, err := apiIdentity(ctx, api)
	if err != nil {
		return nil, fmt.Errorf("VIP API is not ready or authenticated")
	}
	local, err := conn.api(localAPI)
	if err != nil {
		return nil, err
	}
	localIdentity, err := apiIdentity(ctx, local)
	if err != nil {
		return nil, fmt.Errorf("initializer API is not ready or authenticated")
	}
	if identity != localIdentity {
		return nil, fmt.Errorf("VIP belongs to a different cluster")
	}
	return clientcmd.Write(*cfg)
}

func kubeconfigForEndpoint(data []byte, endpoint string) (*clientapi.Config, error) {
	cfg, err := clientcmd.Load(data)
	if err != nil {
		return nil, fmt.Errorf("invalid remote kubeconfig")
	}
	selected := cfg.Contexts[cfg.CurrentContext]
	if selected == nil {
		return nil, fmt.Errorf("remote kubeconfig has no current context")
	}
	cluster, auth := cfg.Clusters[selected.Cluster], cfg.AuthInfos[selected.AuthInfo]
	if cluster == nil || auth == nil || cluster.InsecureSkipTLSVerify || len(cluster.CertificateAuthorityData) == 0 || len(auth.ClientCertificateData) == 0 || len(auth.ClientKeyData) == 0 || auth.Exec != nil || auth.AuthProvider != nil || auth.TokenFile != "" {
		return nil, fmt.Errorf("remote kubeconfig must contain embedded CA and client credentials without plugins")
	}
	return &clientapi.Config{
		CurrentContext: cfg.CurrentContext,
		Contexts: map[string]*clientapi.Context{cfg.CurrentContext: {
			Cluster: selected.Cluster, AuthInfo: selected.AuthInfo, Namespace: selected.Namespace,
		}},
		Clusters: map[string]*clientapi.Cluster{selected.Cluster: {
			Server: endpoint, CertificateAuthorityData: cluster.CertificateAuthorityData,
		}},
		AuthInfos: map[string]*clientapi.AuthInfo{selected.AuthInfo: {
			ClientCertificateData: auth.ClientCertificateData, ClientKeyData: auth.ClientKeyData,
		}},
	}, nil
}

func apiConfig(cfg *clientapi.Config) (*rest.Config, error) {
	options, err := clientcmd.NewDefaultClientConfig(*cfg, &clientcmd.ConfigOverrides{}).ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("invalid remote kubeconfig")
	}
	options.Timeout = 10 * time.Second
	options.Proxy = func(*http.Request) (*url.URL, error) { return nil, nil }
	return options, nil
}
