package secrets

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"k8s.io/apimachinery/pkg/util/validation"
)

// EnvironmentRecords reads explicit secret exports from every Terragrunt unit.
func EnvironmentRecords(ctx context.Context, environment string) (map[string]map[string]string, error) {
	if !filepath.IsLocal(environment) || filepath.Base(environment) != environment {
		return nil, fmt.Errorf("--environment must be an environment directory name")
	}
	root := filepath.Join("infra", environment)
	if _, err := os.Stat(root); err != nil {
		return nil, fmt.Errorf("read environment %s: %w", environment, err)
	}
	units, err := terragruntUnits(root)
	if err != nil {
		return nil, err
	}
	records := make(map[string]map[string]string)
	for _, unit := range units {
		command := exec.CommandContext(ctx, "terragrunt", "run", "--non-interactive", "--no-color", "--no-filters-file", "--tf-forward-stdout", "--", "output", "-json")
		command.Dir = filepath.Join(root, unit)
		output, err := command.Output()
		if err != nil {
			return nil, fmt.Errorf("read Terraform outputs for %s failed", unit)
		}
		if len(bytes.TrimSpace(output)) == 0 {
			continue
		}
		values, err := outputRecords(output, filepath.ToSlash(unit))
		if err != nil {
			return nil, err
		}
		for path, value := range values {
			if _, exists := records[path]; exists {
				return nil, fmt.Errorf("duplicate secret destination %s", path)
			}
			records[path] = value
		}
	}
	return records, nil
}

func terragruntUnits(root string) ([]string, error) {
	var units []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() && strings.HasPrefix(entry.Name(), ".") {
			return filepath.SkipDir
		}
		if entry.IsDir() || entry.Name() != "terragrunt.hcl" {
			return nil
		}
		unit, err := filepath.Rel(root, filepath.Dir(path))
		if err != nil {
			return err
		}
		units = append(units, unit)
		return nil
	})
	return units, err
}

func outputRecords(input []byte, unit string) (map[string]map[string]string, error) {
	var outputs map[string]struct {
		Value json.RawMessage `json:"value"`
	}
	if json.Unmarshal(input, &outputs) != nil || outputs == nil {
		return nil, fmt.Errorf("invalid Terraform output JSON for %s", unit)
	}
	output, exists := outputs["secrets"]
	if !exists {
		return nil, nil
	}
	var exports map[string]map[string]*string
	if json.Unmarshal(output.Value, &exports) != nil || exports == nil {
		return nil, fmt.Errorf("invalid secrets output for %s: expected a map of paths to string fields", unit)
	}
	records := make(map[string]map[string]string, len(exports))
	for path, fields := range exports {
		namespace, _, hasSecret := strings.Cut(path, "/")
		if !hasSecret || len(validation.IsDNS1123Label(namespace)) != 0 || !filepath.IsLocal(path) || filepath.ToSlash(filepath.Clean(path)) != path || strings.Contains(path, "\\") || len(fields) == 0 {
			return nil, fmt.Errorf("invalid secret record in output for %s", unit)
		}
		records[path] = make(map[string]string, len(fields))
		for key, value := range fields {
			if key == "" || value == nil {
				return nil, fmt.Errorf("invalid secret fields in output for %s", unit)
			}
			records[path][key] = *value
		}
	}
	return records, nil
}
