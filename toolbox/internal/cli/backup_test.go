package cli

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestBackupFlags(t *testing.T) {
	for _, args := range [][]string{{"backup", "setup"}, {"backup", "restore", "--pvc", "data"}, {"backup", "setup", "-n", "demo"}, {"backup", "setup", "-n", "", "--pvc", "data"}, {"backup", "setup", "--action", "restore"}} {
		if _, _, err := operationExecute(args...); err == nil {
			t.Fatalf("accepted %v", args)
		}
	}
}

func TestBackupRouting(t *testing.T) {
	operationFake(t, "kubectl", `cat; printf '\n'`)
	for _, tc := range []struct{ action, flag, kind string }{
		{"setup", "--namespace", "ReplicationSource"},
		{"restore", "-n", "ReplicationDestination"},
	} {
		out, stderr, err := operationExecute("backup", tc.action, tc.flag, "media", "--pvc", "videos")
		if err != nil || stderr != "" {
			t.Fatalf("backup command = %q, %v", stderr, err)
		}
		decoder := json.NewDecoder(strings.NewReader(out))
		for _, kind := range []string{"ExternalSecret", tc.kind} {
			var object struct {
				Kind     string
				Metadata struct{ Name, Namespace string }
			}
			if err := decoder.Decode(&object); err != nil || object.Kind != kind || object.Metadata.Namespace != "media" {
				t.Fatalf("CLI backup object = %+v, %v", object, err)
			}
			wantName := "videos"
			if kind == "ExternalSecret" {
				wantName += "-backup-repository"
			}
			if object.Metadata.Name != wantName {
				t.Fatalf("CLI backup object name = %q", object.Metadata.Name)
			}
		}
	}
}
