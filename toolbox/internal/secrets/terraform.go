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
)

// EnvironmentRecords reads every Terragrunt unit's outputs without applying changes.
func EnvironmentRecords(ctx context.Context, environment string) (map[string]map[string]string, error) {
	if !filepath.IsLocal(environment) || filepath.Base(environment) != environment {
		return nil, fmt.Errorf("--environment must be an environment directory name")
	}
	root := filepath.Join("infra", environment)
	if _, err := os.Stat(filepath.Join(root, "root.hcl")); err != nil {
		return nil, fmt.Errorf("read environment %s: %w", environment, err)
	}
	units, err := terragruntUnits(root)
	if err != nil {
		return nil, err
	}
	records := make(map[string]map[string]string)
	for _, unit := range units {
		command := exec.CommandContext(ctx, "terragrunt", "run", "--no-auto-init", "--non-interactive", "--no-color", "--no-filters-file", "--tf-forward-stdout", "--", "output", "-json")
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
	records := make(map[string]map[string]string, len(outputs))
	for name, output := range outputs {
		if name == "" || name == "." || name == ".." || strings.ContainsAny(name, "/\\") {
			return nil, fmt.Errorf("invalid Terraform output name for %s", unit)
		}
		var value any
		decoder := json.NewDecoder(bytes.NewReader(output.Value))
		decoder.UseNumber()
		if decoder.Decode(&value) != nil {
			return nil, fmt.Errorf("invalid Terraform output %s/%s", unit, name)
		}
		text, ok := value.(string)
		if !ok {
			encoded, err := json.Marshal(value)
			if err != nil {
				return nil, fmt.Errorf("encode Terraform output %s/%s failed", unit, name)
			}
			text = string(encoded)
		}
		path := filepath.ToSlash(filepath.Join("infra", unit, name))
		records[path] = map[string]string{"value": text}
	}
	return records, nil
}
