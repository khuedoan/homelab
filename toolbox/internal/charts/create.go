package charts

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

func Create(name string) error {
	if !regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`).MatchString(name) {
		return fmt.Errorf("name must contain only lowercase letters, digits and internal hyphens")
	}
	if err := os.MkdirAll("apps", 0755); err != nil {
		return err
	}
	path := filepath.Join("apps", name)
	if err := os.Mkdir(path, 0755); err != nil {
		return fmt.Errorf("create application directory: %w", err)
	}
	complete := false
	defer func() {
		if !complete {
			_ = os.RemoveAll(path)
		}
	}()
	chart := "apiVersion: v2\nname: CHANGEME\nversion: 0.0.0\ndependencies:\n- name: CHANGEME\n  version: CHANGEME\n  repository: CHANGEME\n"
	if err := os.WriteFile(filepath.Join(path, "Chart.yaml"), []byte(chart), 0644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(path, "values.yaml"), nil, 0644); err != nil {
		return err
	}
	complete = true
	return nil
}
