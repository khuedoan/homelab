package cli

import (
	"testing"
)

func TestEnrollmentArguments(t *testing.T) {
	for _, args := range [][]string{
		{"cluster", "enroll"},
		{"cluster", "enroll", "--environment", "../production"},
		{"cluster", "enroll", "--environment", "staging", "--timeout", "0s"},
		{"cluster", "enroll", "extra", "--environment", "staging"},
	} {
		if _, _, err := operationExecute(args...); err == nil {
			t.Fatalf("accepted %v", args)
		}
	}
}
