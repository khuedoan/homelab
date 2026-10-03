package cli

import (
	"os"
	"testing"

	"github.com/spf13/cobra"
)

func TestRunnerWithoutKubeconfigFlag(t *testing.T) {
	t.Setenv("KUBECONFIG", "original.yaml")
	run := commandRunner(&cobra.Command{})
	for _, want := range []string{"changed.yaml", "unset"} {
		if want == "unset" {
			if err := os.Unsetenv("KUBECONFIG"); err != nil {
				t.Fatal(err)
			}
		} else {
			t.Setenv("KUBECONFIG", want)
		}
		output, err := run.Output(t.Context(), "sh", "-c", `printf '%s' "${KUBECONFIG-unset}"`)
		if err != nil || string(output) != want {
			t.Fatalf("no-flag inheritance = %q, %v", output, err)
		}
	}
}
