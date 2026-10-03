package identity

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/khuedoan/homelab/toolbox/internal/process"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ingressHost(ctx context.Context, run process.Runner, namespace, name string) (string, error) {
	output, err := run.Output(ctx, "kubectl", "get", "ingress", "--namespace", namespace, name, "--output=jsonpath={.spec.rules[0].host}")
	if err != nil {
		return "", err
	}
	host := strings.TrimSpace(string(output))
	if host == "" {
		return "", fmt.Errorf("ingress %s/%s has no host", namespace, name)
	}
	return host, nil
}

func readSecret(ctx context.Context, run process.Runner, namespace, name string) (map[string]string, error) {
	out, err := run.PrivateOutput(ctx, "kubectl", "get", "secret", "--namespace", namespace, name, "--output=json")
	if err != nil {
		return nil, err
	}
	var secret struct {
		Data map[string]string `json:"data"`
	}
	if json.Unmarshal(out, &secret) != nil {
		return nil, fmt.Errorf("secret %s/%s has invalid JSON", namespace, name)
	}
	values := make(map[string]string)
	for key, value := range secret.Data {
		decoded, err := base64.StdEncoding.DecodeString(value)
		if err != nil {
			return nil, fmt.Errorf("secret %s/%s has invalid %s", namespace, name, key)
		}
		values[key] = string(decoded)
	}
	return values, nil
}

func applySecret(ctx context.Context, run process.Runner, name string, values map[string]string) error {
	object := corev1.Secret{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Secret"},
		ObjectMeta: metav1.ObjectMeta{
			Name: name, Namespace: "global-secrets",
		},
		Type:       corev1.SecretTypeOpaque,
		StringData: values,
	}
	data, err := json.Marshal(object)
	if err != nil {
		return err
	}
	// kubectl can echo input on failure, so suppress its output for Secret writes.
	p := run.Command(ctx, "kubectl", "apply", "--filename=-")
	p.Stdin = bytes.NewReader(data)
	if err := p.Run(); err != nil {
		return fmt.Errorf("apply secret %s: %w", name, err)
	}
	return nil
}
