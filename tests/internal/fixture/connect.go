package fixture

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/core/v2/logger"
	"github.com/gruntwork-io/terratest/modules/core/v2/shell"
	"github.com/khuedoan/homelab/tests/internal/testenv"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// Cluster binds the validated target to clients using only its exported kubeconfig.
type Cluster struct {
	Target  testenv.Target
	Client  *kubernetes.Clientset
	Dynamic dynamic.Interface
}

// Connect verifies host identities over trusted SSH before exporting a private kubeconfig.
// The export must select the target VIP, verify API TLS, and avoid credential plugins.
func Connect(t *testing.T, target testenv.Target) Cluster {
	t.Helper()
	checkHosts(t, target)
	work := t.TempDir()
	toolbox := os.Getenv("TOOLBOX_BIN")
	if toolbox == "" {
		toolbox = filepath.Join(work, "toolbox")
		run(t, target, "build toolbox", "go", "-C", filepath.Join(target.Root, "toolbox"), "build", "-o", toolbox, ".")
	} else {
		require.True(t, filepath.IsAbs(toolbox), "TOOLBOX_BIN must be an absolute path")
	}
	kubeconfig := filepath.Join(work, "kubeconfig.yaml")
	run(t, target, "export kubeconfig", toolbox, "cluster", "kubeconfig", "--environment", target.Environment, "--output", kubeconfig, "--timeout", "30s")
	info, err := os.Stat(kubeconfig)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0600), info.Mode().Perm())
	raw, err := clientcmd.LoadFromFile(kubeconfig)
	if err != nil {
		t.Fatal("exported kubeconfig is invalid")
	}
	require.Len(t, raw.Clusters, 1)
	require.Len(t, raw.AuthInfos, 1)
	for _, auth := range raw.AuthInfos {
		require.True(t, auth.Exec == nil && auth.AuthProvider == nil, "export must not use credential plugins")
	}
	for _, endpoint := range raw.Clusters {
		require.Equal(t, "https://"+target.Cluster.VIP+":6443", endpoint.Server)
		require.False(t, endpoint.InsecureSkipTLSVerify)
	}
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		t.Fatal("cannot load the explicit test kubeconfig")
	}
	config.Timeout = 10 * time.Second
	client, err := kubernetes.NewForConfig(config)
	require.NoError(t, err)
	dynamicClient, err := dynamic.NewForConfig(config)
	require.NoError(t, err)
	return Cluster{Target: target, Client: client, Dynamic: dynamicClient}
}

func run(t *testing.T, target testenv.Target, action, executable string, args ...string) string {
	t.Helper()
	output, err := shell.RunCommandContextAndGetOutputE(t, t.Context(), &shell.Command{
		Command: executable, Args: args, WorkingDir: target.Root, Logger: logger.Discard,
	})
	if err != nil {
		t.Fatalf("%s failed; command output withheld to protect credentials", action)
	}
	return strings.TrimSpace(output)
}

// SSH runs a bounded root command on a validated inventory host using trusted host keys.
// Command failures withhold output to avoid exposing credentials.
func SSH(t *testing.T, target testenv.Target, name, command string) string {
	t.Helper()
	args := []string{
		"-F", "/dev/null",
		"-o", "BatchMode=yes",
		"-o", "StrictHostKeyChecking=yes",
		"-o", "ConnectTimeout=5",
		"-o", "ConnectionAttempts=1",
	}
	if os.Getenv("SSH_AUTH_SOCK") == "" {
		key := os.Getenv("SSH_KEY")
		if key == "" {
			home, err := os.UserHomeDir()
			require.NoError(t, err)
			key = filepath.Join(home, ".ssh", "id_ed25519")
		}
		args = append(args, "-i", key, "-o", "IdentitiesOnly=yes")
	}
	node := target.Hosts[name]
	args = append(args, "root@"+node.IP, "timeout 20s sh -c '"+strings.ReplaceAll(command, "'", "'\\''")+"'")
	return run(t, target, "SSH check on "+name+" ("+node.IP+")", "ssh", args...)
}
