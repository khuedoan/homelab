package cluster

import (
	"context"
	"fmt"
	"net/netip"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	coreclient "k8s.io/client-go/kubernetes/typed/core/v1"
)

const localAPI = "https://127.0.0.1:6443"

func (remote sshServers) checkHost(ctx context.Context, node host) error {
	conn, err := remote.connect(ctx, node)
	if err != nil {
		return err
	}
	defer conn.close()
	return conn.checkHost(node)
}

func (conn *connection) checkHost(node host) error {
	name, err := readFile(conn.files, "/etc/hostname")
	if err != nil {
		return err
	}
	if strings.TrimSpace(name) != node.name {
		return fmt.Errorf("management address for %s belongs to a different host", node.name)
	}
	return nil
}

func (conn *connection) api(endpoint string) (*coreclient.CoreV1Client, error) {
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
	options.Dial = conn.ssh.DialContext
	return coreclient.NewForConfig(options)
}

func apiIdentity(ctx context.Context, api coreclient.CoreV1Interface) (string, error) {
	if err := api.RESTClient().Get().AbsPath("/readyz").Do(ctx).Error(); err != nil {
		return "", err
	}
	namespace, err := api.Namespaces().Get(ctx, "kube-system", metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	if namespace.UID == "" {
		return "", fmt.Errorf("cluster has no kube-system identity")
	}
	return string(namespace.UID), nil
}

func (remote sshServers) source(ctx context.Context, seed host, vip netip.Addr) (credentials, error) {
	conn, err := remote.connect(ctx, seed)
	if err != nil {
		return credentials{}, err
	}
	defer conn.close()
	local, err := conn.api(localAPI)
	if err != nil {
		return credentials{}, err
	}
	identity, err := apiIdentity(ctx, local)
	if err != nil {
		return credentials{}, err
	}
	api, err := conn.api("https://" + vip.String() + ":6443")
	if err != nil {
		return credentials{}, err
	}
	vipIdentity, err := apiIdentity(ctx, api)
	if err != nil {
		return credentials{}, err
	}
	if vipIdentity != identity {
		return credentials{}, fmt.Errorf("VIP belongs to a different cluster")
	}
	token, err := readFile(conn.files, "/var/lib/rancher/k3s/server/token")
	if err != nil {
		return credentials{}, err
	}
	token = strings.TrimSpace(token)
	if !validToken(token) {
		return credentials{}, fmt.Errorf("initializer did not provide a secure k3s server token")
	}
	return credentials{token, identity}, nil
}

func (remote sshServers) ready(ctx context.Context, node host, clusterID string) error {
	conn, err := remote.connect(ctx, node)
	if err != nil {
		return err
	}
	defer conn.close()
	api, err := conn.api(localAPI)
	if err != nil {
		return err
	}
	uid, err := apiIdentity(ctx, api)
	if err != nil {
		return err
	}
	if uid != clusterID {
		return fmt.Errorf("node %s is not a member of the intended cluster", node.name)
	}
	return nil
}

func (remote sshServers) verify(ctx context.Context, cfg config, source credentials) error {
	current, err := remote.source(ctx, cfg.seed, cfg.vip)
	if err != nil {
		return err
	}
	if current != source {
		return fmt.Errorf("initializer identity or credentials changed")
	}
	localIDs := make(map[string]string)
	for _, node := range append([]host{cfg.seed}, cfg.joiners...) {
		uid, err := remote.nodeUID(ctx, node, source.clusterID)
		if err != nil {
			return err
		}
		localIDs[node.name] = uid
	}
	conn, err := remote.connect(ctx, cfg.seed)
	if err != nil {
		return err
	}
	defer conn.close()
	api, err := conn.api(localAPI)
	if err != nil {
		return err
	}
	nodes, err := api.Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	return verifyNodes(nodes.Items, localIDs)
}

func (remote sshServers) nodeUID(ctx context.Context, node host, clusterID string) (string, error) {
	conn, err := remote.connect(ctx, node)
	if err != nil {
		return "", err
	}
	defer conn.close()
	api, err := conn.api(localAPI)
	if err != nil {
		return "", err
	}
	uid, err := apiIdentity(ctx, api)
	if err != nil {
		return "", err
	}
	if uid != clusterID {
		return "", fmt.Errorf("node %s belongs to another cluster", node.name)
	}
	member, err := api.Nodes().Get(ctx, node.name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	return string(member.UID), nil
}

func verifyNodes(nodes []corev1.Node, localIDs map[string]string) error {
	for name, uid := range localIDs {
		found := false
		for _, node := range nodes {
			if node.Name == name && uid != "" && string(node.UID) == uid && readyNode(node) {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("node %s is not a Ready control-plane member", name)
		}
	}
	return nil
}

func readyNode(node corev1.Node) bool {
	if _, server := node.Labels["node-role.kubernetes.io/control-plane"]; !server {
		return false
	}
	for _, condition := range node.Status.Conditions {
		if condition.Type == corev1.NodeReady && condition.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}
