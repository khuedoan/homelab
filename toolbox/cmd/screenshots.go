package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func newScreenshotsCmd() *cobra.Command {
	var profile, output string
	cmd := &cobra.Command{
		Use: "capture", Short: "Take application screenshots with headless Firefox",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			directory, err := filepath.Abs(output)
			if err != nil {
				return err
			}
			if err := os.MkdirAll(directory, 0755); err != nil {
				return err
			}
			if profile != "" {
				profile, err = filepath.Abs(profile)
				if err != nil {
					return err
				}
			}
			apps := []struct{ name, url string }{
				{"home", "https://home.khuedoan.com"},
				{"gitea", "https://git.khuedoan.com/ops/homelab"},
				{"argocd", "https://argocd.khuedoan.com/applications/root"},
				{"matrix", "https://chat.khuedoan.com/#/room/#random:matrix.khuedoan.com"},
				{"grafana", "https://grafana.khuedoan.com/d/efa86fd1d0c121a26444b636a3f509a8/kubernetes-compute-resources-cluster"},
			}
			for _, app := range apps {
				path := filepath.Join(directory, app.name+".png")
				args := []string{"--headless", "--window-size", "1920,1080"}
				if profile != "" {
					args = append(args, "--profile", profile)
				}
				args = append(args, "--screenshot", path, app.url)
				if err := runCommand(cmd, nil, "firefox", args...); err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Screenshot saved to %s\n", path)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&profile, "profile", "", "Firefox profile directory to use")
	cmd.Flags().StringVar(&output, "output", ".", "Screenshot output directory")
	return cmd
}
