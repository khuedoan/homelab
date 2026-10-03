package process

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestCommandEnvironment(t *testing.T) {
	t.Setenv("KUBECONFIG", "original.yaml")
	t.Setenv("PRESERVED_ENV", "untouched")
	for _, tc := range []struct{ selection, want string }{
		{"", "changed.yaml|untouched"},
		{"selected.yaml", "selected.yaml|untouched"},
	} {
		run := New(tc.selection, nil, nil, nil)
		t.Setenv("KUBECONFIG", "changed.yaml")
		output, err := run.Output(t.Context(), "sh", "-c", `printf '%s|%s' "$KUBECONFIG" "$PRESERVED_ENV"`)
		if err != nil || string(output) != tc.want {
			t.Fatalf("child environment = %q, %v", output, err)
		}
		if os.Getenv("KUBECONFIG") != "changed.yaml" {
			t.Fatal("changed the parent environment")
		}
	}
}

func TestRunStreamsAndInput(t *testing.T) {
	for _, tc := range []struct {
		input io.Reader
		want  string
	}{
		{nil, "configured input"},
		{strings.NewReader("explicit input"), "explicit input"},
		{strings.NewReader(""), ""},
	} {
		var stdout, stderr bytes.Buffer
		run := New("", strings.NewReader("configured input"), &stdout, &stderr)
		err := run.Run(t.Context(), tc.input, "sh", "-c", `cat; printf diagnostic >&2`)
		if err != nil || stdout.String() != tc.want || stderr.String() != "diagnostic" {
			t.Fatalf("streams = %q, %q, %v", &stdout, &stderr, err)
		}
	}
	var stdout, stderr bytes.Buffer
	run := New("", strings.NewReader("configured input"), &stdout, &stderr)
	output, err := run.Output(t.Context(), "sh", "-c", `cat; printf captured; printf diagnostic >&2`)
	if err != nil || string(output) != "captured" || stdout.Len() != 0 || stderr.String() != "diagnostic" {
		t.Fatalf("captured output = %q, %q, %q, %v", output, &stdout, &stderr, err)
	}
	err = run.Run(t.Context(), nil, "sh", "-c", "exit 17")
	if err == nil || err.Error() != "sh: exit status 17" {
		t.Fatalf("run failure = %v", err)
	}
}

func TestPrivateOutput(t *testing.T) {
	var stdout, stderr bytes.Buffer
	run := New("", strings.NewReader("must not be consumed"), &stdout, &stderr)
	output, err := run.PrivateOutput(t.Context(), "sh", "-c", `cat; printf private-value; printf private-diagnostic >&2`)
	if err != nil || string(output) != "private-value" || stdout.Len() != 0 || stderr.Len() != 0 {
		t.Fatalf("private output = %q, %q, %q, %v", output, &stdout, &stderr, err)
	}
	output, err = run.PrivateOutput(t.Context(), "sh", "-c", `printf private-value; printf private-diagnostic >&2; exit 17`)
	var exit *exec.ExitError
	if err == nil || err.Error() != "sh failed: exit status 17" || !errors.As(err, &exit) {
		t.Fatalf("private failure = %v", err)
	}
	if output != nil || stdout.Len() != 0 || stderr.Len() != 0 || len(exit.Stderr) != 0 {
		t.Fatal("private failure retained or forwarded credentials")
	}
}

func TestCanceledCommand(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	output, err := New("", nil, nil, nil).Output(ctx, "sh", "-c", "printf unexpected")
	if !errors.Is(err, context.Canceled) || len(output) != 0 {
		t.Fatalf("canceled command = %q, %v", output, err)
	}
}
