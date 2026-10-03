package backup

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/khuedoan/homelab/toolbox/internal/process"
)

func fakeKubectl(t *testing.T, script string) {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "kubectl"), []byte("#!/bin/sh\n"+script), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+":"+os.Getenv("PATH"))
}

func TestBackupManifests(t *testing.T) {
	file := filepath.Join(t.TempDir(), "objects")
	t.Setenv("OBJECTS", file)
	fakeKubectl(t, `[ "$*" = 'apply --filename=-' ] || exit 99
cat >> "$OBJECTS"; printf '\n' >> "$OBJECTS"
if [ "$FAIL" = yes ]; then exit 17; fi
printf 'applied\n'`)
	secret := `{"apiVersion":"external-secrets.io/v1beta1","kind":"ExternalSecret","metadata":{"name":"data-backup-repository","namespace":"demo","annotations":{"app.kubernetes.io/managed-by":"toolbox"}},"spec":{"secretStoreRef":{"kind":"ClusterSecretStore","name":"global-secrets"},"data":[{"remoteRef":{"key":"external","property":"restic-s3-bucket"},"secretKey":"restic_s3_bucket"},{"remoteRef":{"key":"external","property":"restic-s3-access-key"},"secretKey":"restic_s3_access_key"},{"remoteRef":{"key":"external","property":"restic-s3-secret-key"},"secretKey":"restic_s3_secret_key"},{"remoteRef":{"key":"external","property":"restic-password"},"secretKey":"restic_password"}],"target":{"template":{"data":{"RESTIC_REPOSITORY":"s3:{{ .restic_s3_bucket }}/demo/data","RESTIC_PASSWORD":"{{ .restic_password }}","AWS_ACCESS_KEY_ID":"{{ .restic_s3_access_key }}","AWS_SECRET_ACCESS_KEY":"{{ .restic_s3_secret_key }}"}}}}}`
	for _, tc := range []struct{ action, kind, spec string }{
		{"setup", "ReplicationSource", `{"sourcePVC":"data","trigger":{"schedule":"*/30 * * * *"},"restic":{"pruneIntervalDays":14,"repository":"data-backup-repository","retain":{"hourly":6,"daily":5,"weekly":4,"monthly":2,"yearly":1},"copyMethod":"Snapshot"}}`},
		{"restore", "ReplicationDestination", `{"trigger":{"manual":"restore-once"},"restic":{"repository":"data-backup-repository","destinationPVC":"data","copyMethod":"Direct"}}`},
	} {
		if err := os.WriteFile(file, nil, 0600); err != nil {
			t.Fatal(err)
		}
		var out bytes.Buffer
		run := process.New("", nil, &out, nil)
		volume := PVC{Namespace: "demo", Name: "data"}
		operation := Setup
		if tc.action == "restore" {
			operation = Restore
		}
		if err := operation(t.Context(), run, volume); err != nil || out.String() != "applied\napplied\n" {
			t.Fatalf("%q, %v", &out, err)
		}
		data, _ := os.ReadFile(file)
		lines := strings.Split(strings.TrimSpace(string(data)), "\n")
		if len(lines) != 2 {
			t.Fatalf("objects: %s", data)
		}
		volsync := `{"apiVersion":"volsync.backube/v1alpha1","kind":"` + tc.kind + `","metadata":{"name":"data","namespace":"demo","annotations":{"app.kubernetes.io/managed-by":"toolbox"}},"spec":` + tc.spec + `}`
		for i, expected := range []string{secret, volsync} {
			var got, want any
			if err := json.Unmarshal([]byte(lines[i]), &got); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal([]byte(expected), &want); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("manifest mismatch\ngot %s\nwant %s", lines[i], expected)
			}
		}
	}
	t.Setenv("FAIL", "yes")
	if err := os.WriteFile(file, nil, 0600); err != nil {
		t.Fatal(err)
	}
	if err := Setup(t.Context(), process.New("", nil, nil, nil), PVC{Namespace: "demo", Name: "data"}); err == nil {
		t.Fatal("ignored apply failure")
	}
	data, _ := os.ReadFile(file)
	if strings.Count(string(data), "\n") != 1 {
		t.Fatal("continued after failed ExternalSecret")
	}
}
