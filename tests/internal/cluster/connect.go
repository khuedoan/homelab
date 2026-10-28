package cluster

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/core/v2/logger"
	"github.com/gruntwork-io/terratest/modules/core/v2/shell"
	"github.com/khuedoan/homelab/tests/internal/inventory"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Cluster binds the validated target to clients using only its exported kubeconfig.
type Cluster struct {
	Target     inventory.Target
	Client     *kubernetes.Clientset
	Kubeconfig string
	config     *rest.Config
}

// RESTConfig returns a copy of the validated configuration for upstream clients.
func (cluster Cluster) RESTConfig() *rest.Config {
	return rest.CopyConfig(cluster.config)
}

// Connect exports a private kubeconfig using the lab CLI and creates explicit clients.
func Connect(ctx context.Context, t testing.TB, target inventory.Target) Cluster {
	t.Helper()
	work := t.TempDir()
	lab := os.Getenv("LAB_BIN")
	if lab == "" {
		lab = filepath.Join(work, "lab")
		run(ctx, t, target, "build lab", "go", "-C", filepath.Join(target.Root, "cli"), "build", "-mod=readonly", "-o", lab, "./cmd/lab")
	} else {
		require.True(t, filepath.IsAbs(lab), "LAB_BIN must be an absolute path")
	}
	kubeconfig := filepath.Join(work, "kubeconfig.yaml")
	run(ctx, t, target, "export kubeconfig", lab, "cluster", "kubeconfig", "--environment", target.Environment, "--output", kubeconfig, "--timeout", "30s")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		t.Fatal("cannot load the explicit test kubeconfig")
	}
	require.Equal(t, "https://"+target.Cluster.VIP+":6443", config.Host)
	config.Timeout = 10 * time.Second
	client, err := kubernetes.NewForConfig(config)
	require.NoError(t, err)
	return Cluster{Target: target, Client: client, Kubeconfig: kubeconfig, config: config}
}

func run(ctx context.Context, t testing.TB, target inventory.Target, action, executable string, args ...string) string {
	t.Helper()
	output, err := shell.RunCommandContextAndGetOutputE(t, ctx, &shell.Command{
		Command: executable, Args: args, WorkingDir: target.Root, Logger: logger.Discard,
	})
	if err != nil {
		t.Fatalf("%s failed; command output withheld to protect credentials", action)
	}
	return strings.TrimSpace(output)
}

// SSH runs a bounded root command on a validated inventory host using trusted host keys.
// Command failures withhold output to avoid exposing credentials.
func SSH(ctx context.Context, t testing.TB, target inventory.Target, name, command string) string {
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
	return run(ctx, t, target, "SSH check on "+name+" ("+node.IP+")", "ssh", args...)
}
