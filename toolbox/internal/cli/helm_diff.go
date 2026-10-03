package cli

import (
	"github.com/khuedoan/homelab/toolbox/internal/charts"
	"github.com/spf13/cobra"
)

func newHelmDiffCmd() *cobra.Command {
	var repository, source, target, subpath string
	cmd := &cobra.Command{
		Use: "diff", Short: "Compare Helm charts between Git revisions",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return charts.Diff(cmd.Context(), commandRunner(cmd), charts.DiffOptions{Repository: repository, Source: source, Target: target, Subpath: subpath})
		},
	}
	cmd.Flags().StringVar(&repository, "repository", "", "Repository to clone")
	cmd.Flags().StringVar(&source, "source", "", "Source Git revision")
	cmd.Flags().StringVar(&target, "target", "", "Target Git revision")
	cmd.Flags().StringVar(&subpath, "subpath", "", "Repository directory containing charts")
	for _, flag := range []string{"repository", "source", "target", "subpath"} {
		_ = cmd.MarkFlagRequired(flag)
	}
	return cmd
}
