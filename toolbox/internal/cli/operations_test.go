package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestOperationCommands(t *testing.T) {
	operationFake(t, "kubectl", `if [ "$FAIL" = yes ]; then echo failed >&2; exit 17; fi
case "$*" in
'get applicationsets --namespace argocd') printf 'APPLICATIONSETS\n';;
'get applications --namespace argocd') printf 'APPLICATIONS\n';;
'get ingress --all-namespaces') printf 'INGRESSES\n';;
'get ingress --all-namespaces --no-headers --output custom-columns=ADDRESS:.status.loadBalancer.ingress[0].ip,HOST:.spec.rules[0].host') printf '192.0.2.1 home.example.com\n';;
'-n argocd get secret argocd-initial-admin-secret -o jsonpath={.data.password}') printf '%s' "${PASSWORD-c2VjcmV0}";;
'exec -it -n kanidm statefulset/kanidm -- kanidmd recover-account admin') printf 'recovered\n';;
'-n wireguard exec -it deployment/wireguard -- /app/show-peer phone') printf 'QR\n';;
'-n wireguard exec -it deployment/wireguard -- cat /config/peer_phone/peer_phone.conf') printf '[Interface]\n';;
*) exit 99;; esac`)
	for _, tc := range []struct {
		args    []string
		want    string
		warning bool
	}{
		{[]string{"status"}, "APPLICATIONSETS\nAPPLICATIONS\nINGRESSES\n", false},
		{[]string{"dns", "list"}, "192.0.2.1 home.example.com\n", false},
		{[]string{"argocd", "admin-password"}, "secret", true},
		{[]string{"users", "reset-password", "admin"}, "recovered\n", true},
		{[]string{"wireguard", "config", "phone"}, "QR\n[Interface]\n", false},
	} {
		out, stderr, err := operationExecute(tc.args...)
		if err != nil || out != tc.want || strings.Contains(stderr, "WARNING:") != tc.warning {
			t.Fatalf("%v: %q, %q, %v", tc.args, out, stderr, err)
		}
		t.Setenv("FAIL", "yes")
		if _, _, err := operationExecute(tc.args...); err == nil {
			t.Fatalf("%v ignored failure", tc.args)
		}
		t.Setenv("FAIL", "")
	}
	t.Setenv("PASSWORD", "not-base64")
	if _, _, err := operationExecute("argocd", "admin-password"); err == nil {
		t.Fatal("accepted invalid base64")
	}
}

func TestOnboardUser(t *testing.T) {
	operationFake(t, "kubectl", `printf kanidm.example.com`)
	operationFake(t, "kanidm", `printf '<%s>' "$@"; printf '\n'; if [ "$2" = "$FAIL_STEP" ]; then exit 17; fi`)
	want := "<person><create><alice><Alice Example><--url><https://kanidm.example.com><--name><idm_admin>\n<person><update><alice><--url><https://kanidm.example.com><--name><idm_admin><--mail><alice@example.com>\n<group><add-members><editor><alice><--url><https://kanidm.example.com><--name><idm_admin>\n<person><credential><create-reset-token><alice><--url><https://kanidm.example.com><--name><idm_admin>\n"
	out, _, err := operationExecute("users", "create", "alice", "Alice Example", "alice@example.com")
	if err != nil || out != want {
		t.Fatalf("%q, %v", out, err)
	}
	t.Setenv("FAIL_STEP", "update")
	out, _, err = operationExecute("users", "create", "alice", "Alice Example", "alice@example.com")
	if err == nil || strings.Contains(out, "<group>") {
		t.Fatalf("did not stop: %q, %v", out, err)
	}
	operationFake(t, "kubectl", `exit 18`)
	if out, _, err := operationExecute("users", "create", "alice", "Alice Example", "alice@example.com"); err == nil || out != "" {
		t.Fatalf("ingress failure: %q %v", out, err)
	}
}

func TestOperationKubeconfig(t *testing.T) {
	operationFake(t, "kubectl", `printf '%s\n' "$KUBECONFIG"`)
	for _, tc := range []struct {
		env, want string
		flags     []string
	}{
		{"", "infra/kubeconfig.yaml", nil},
		{"env.yaml", "env.yaml", nil},
		{"env.yaml", "flag.yaml", []string{"--kubeconfig", "flag.yaml"}},
		{"env.yaml", "infra/kubeconfig.yaml", []string{"--kubeconfig="}},
	} {
		t.Setenv("KUBECONFIG", tc.env)
		args := append([]string{"dns", "list"}, tc.flags...)
		out, _, err := operationExecute(args...)
		if err != nil || out != tc.want+"\n" {
			t.Fatalf("%q, %v", out, err)
		}
		if os.Getenv("KUBECONFIG") != tc.env {
			t.Fatal("command changed the parent kubeconfig environment")
		}
	}
}

func TestIndependentCommandKubeconfigs(t *testing.T) {
	operationFake(t, "kubectl", `printf '%s\n' "$KUBECONFIG"`)
	t.Setenv("KUBECONFIG", "ambient.yaml")
	first, second := newRootCmd(), newRootCmd()
	first.SetArgs([]string{"dns", "list", "--kubeconfig", "first.yaml"})
	second.SetArgs([]string{"dns", "list", "--kubeconfig", "second.yaml"})
	var out bytes.Buffer
	first.SetOut(&out)
	second.SetOut(&out)
	if err := first.Execute(); err != nil {
		t.Fatal(err)
	}
	if err := second.Execute(); err != nil {
		t.Fatal(err)
	}
	if err := first.Execute(); err != nil {
		t.Fatal(err)
	}
	if out.String() != "first.yaml\nsecond.yaml\nfirst.yaml\n" || os.Getenv("KUBECONFIG") != "ambient.yaml" {
		t.Fatalf("commands shared kubeconfig state: %q", out.String())
	}
}

func TestAppScaffold(t *testing.T) {
	t.Chdir(t.TempDir())
	if _, _, err := operationExecute("apps", "create", "example"); err != nil {
		t.Fatal(err)
	}
	chart, err := os.ReadFile("apps/example/Chart.yaml")
	want := "apiVersion: v2\nname: CHANGEME\nversion: 0.0.0\ndependencies:\n- name: CHANGEME\n  version: CHANGEME\n  repository: CHANGEME\n"
	if err != nil || string(chart) != want {
		t.Fatalf("chart: %q, %v", chart, err)
	}
	values, err := os.ReadFile("apps/example/values.yaml")
	if err != nil || len(values) != 0 {
		t.Fatalf("values: %q, %v", values, err)
	}
	for _, name := range []string{"example", "../escape", "/absolute", "nested/app", ".", "UPPER", "bad name"} {
		if _, _, err := operationExecute("apps", "create", name); err == nil {
			t.Fatalf("accepted %q", name)
		}
	}
	unchanged, _ := os.ReadFile("apps/example/Chart.yaml")
	if string(unchanged) != want {
		t.Fatal("overwrote existing chart")
	}
}

func TestOperationArguments(t *testing.T) {
	for _, args := range [][]string{{"status", "extra"}, {"dns", "list", "extra"}, {"argocd", "admin-password", "extra"}, {"users", "reset-password"}, {"users", "create", "alice"}, {"wireguard", "config"}, {"apps", "create"}} {
		if _, _, err := operationExecute(args...); err == nil {
			t.Fatalf("accepted %v", args)
		}
	}
}
