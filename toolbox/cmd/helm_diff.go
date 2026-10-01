package cmd

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

func newHelmDiffCmd() *cobra.Command {
	var repository, source, target, subpath string
	cmd := &cobra.Command{
		Use: "diff", Short: "Compare Helm charts between Git revisions",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clean := filepath.Clean(subpath)
			if filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
				return fmt.Errorf("subpath must stay within the repository")
			}
			tmp, err := os.MkdirTemp("", "toolbox-helm-diff-")
			if err != nil {
				return err
			}
			defer os.RemoveAll(tmp)
			roots := []string{filepath.Join(tmp, "target"), filepath.Join(tmp, "source")}
			for i, ref := range []string{target, source} {
				if err := runCommand(cmd, nil, "git", "clone", "--no-checkout", "--depth", "1", "--", repository, roots[i]); err != nil {
					return err
				}
				if err := runCommand(cmd, nil, "git", "-C", roots[i], "fetch", "--depth", "1", "origin", ref); err != nil {
					return err
				}
				if err := runCommand(cmd, nil, "git", "-C", roots[i], "checkout", "--detach", "FETCH_HEAD"); err != nil {
					return err
				}
				roots[i] = filepath.Join(roots[i], clean)
			}
			charts := make(map[string]bool)
			for _, root := range roots {
				found, err := helmChartDirectories(root)
				if err != nil {
					return err
				}
				for _, chart := range found {
					charts[chart] = true
				}
			}
			names := make([]string, 0, len(charts))
			for chart := range charts {
				names = append(names, chart)
			}
			sort.Strings(names)
			for _, chart := range names {
				targetFiles, err := helmChartFiles(filepath.Join(roots[0], chart))
				if err != nil {
					return err
				}
				sourceFiles, err := helmChartFiles(filepath.Join(roots[1], chart))
				if err != nil {
					return err
				}
				if reflect.DeepEqual(targetFiles, sourceFiles) {
					continue
				}
				manifests := []string{filepath.Join(tmp, "target.yaml"), filepath.Join(tmp, "source.yaml")}
				for i, files := range []map[string]string{targetFiles, sourceFiles} {
					var rendered []byte
					if files != nil {
						path := filepath.Join(roots[i], chart)
						if err := runCommand(cmd, nil, "helm", "dependency", "update", path); err != nil {
							return err
						}
						rendered, err = commandOutput(cmd, "helm", "template", "--namespace", filepath.Base(path), filepath.Base(path), path)
						if err != nil {
							return err
						}
					}
					if err := os.WriteFile(manifests[i], rendered, 0600); err != nil {
						return err
					}
				}
				if err := runCommand(cmd, bytes.NewReader(nil), "dyff", "between", "--omit-header", "--use-go-patch-style", "--color=on", "--truecolor=off", manifests[0], manifests[1]); err != nil {
					return err
				}
			}
			return nil
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

// Stop at each chart root so vendored dependencies aren't separate releases.
func helmChartDirectories(root string) ([]string, error) {
	var charts []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if os.IsNotExist(err) && path == root {
			return nil
		}
		if err != nil {
			return err
		}
		if !entry.IsDir() {
			return nil
		}
		if entry.Name() == ".git" {
			return filepath.SkipDir
		}
		info, err := os.Stat(filepath.Join(path, "Chart.yaml"))
		if os.IsNotExist(err) {
			return nil
		}
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		charts = append(charts, rel)
		return filepath.SkipDir
	})
	return charts, err
}

func helmChartFiles(root string) (map[string]string, error) {
	if _, err := os.Stat(filepath.Join(root, "Chart.yaml")); os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	files := make(map[string]string)
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if entry.Type()&os.ModeSymlink != 0 {
			link, err := os.Readlink(path)
			files[rel] = "symlink:" + link
			return err
		}
		data, err := os.ReadFile(path)
		files[rel] = "file:" + string(data)
		return err
	})
	return files, err
}
