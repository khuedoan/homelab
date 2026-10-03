package charts

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/khuedoan/homelab/toolbox/internal/process"
)

type DiffOptions struct {
	Repository string
	Source     string
	Target     string
	Subpath    string
}

func Diff(ctx context.Context, run process.Runner, options DiffOptions) error {
	clean := filepath.Clean(options.Subpath)
	if filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return fmt.Errorf("subpath must stay within the repository")
	}
	tmp, err := os.MkdirTemp("", "toolbox-helm-diff-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	type revision struct {
		ref, root, manifest string
	}
	target := revision{options.Target, filepath.Join(tmp, "target"), filepath.Join(tmp, "target.yaml")}
	source := revision{options.Source, filepath.Join(tmp, "source"), filepath.Join(tmp, "source.yaml")}
	for _, side := range []*revision{&target, &source} {
		if err := run.Run(ctx, nil, "git", "clone", "--no-checkout", "--depth", "1", "--", options.Repository, side.root); err != nil {
			return err
		}
		if err := run.Run(ctx, nil, "git", "-C", side.root, "fetch", "--depth", "1", "origin", side.ref); err != nil {
			return err
		}
		if err := run.Run(ctx, nil, "git", "-C", side.root, "checkout", "--detach", "FETCH_HEAD"); err != nil {
			return err
		}
		side.root = filepath.Join(side.root, clean)
	}
	charts := make(map[string]bool)
	for _, side := range []revision{target, source} {
		found, err := topLevelCharts(side.root)
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
		targetFiles, err := helmChartFiles(filepath.Join(target.root, chart))
		if err != nil {
			return err
		}
		sourceFiles, err := helmChartFiles(filepath.Join(source.root, chart))
		if err != nil {
			return err
		}
		if reflect.DeepEqual(targetFiles, sourceFiles) {
			continue
		}
		for _, side := range []struct {
			revision
			files map[string]string
		}{{target, targetFiles}, {source, sourceFiles}} {
			var rendered []byte
			if side.files != nil {
				path := filepath.Join(side.root, chart)
				if err := run.Run(ctx, nil, "helm", "dependency", "update", path); err != nil {
					return err
				}
				rendered, err = run.Output(ctx, "helm", "template", "--namespace", filepath.Base(path), filepath.Base(path), path)
				if err != nil {
					return err
				}
			}
			if err := os.WriteFile(side.manifest, rendered, 0600); err != nil {
				return err
			}
		}
		if err := run.Run(ctx, bytes.NewReader(nil), "dyff", "between", "--omit-header", "--use-go-patch-style", "--color=on", "--truecolor=off", target.manifest, source.manifest); err != nil {
			return err
		}
	}
	return nil
}

func topLevelCharts(root string) ([]string, error) {
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
