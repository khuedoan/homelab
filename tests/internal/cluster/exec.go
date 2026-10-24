package cluster

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
)

// Exec runs a bounded command and returns stdout without trimming or mixing stderr.
func (cluster Cluster) Exec(ctx context.Context, namespace, pod string, command ...string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	request := cluster.Client.CoreV1().RESTClient().Post().Resource("pods").Namespace(namespace).Name(pod).SubResource("exec").VersionedParams(&corev1.PodExecOptions{
		Container: "main", Command: command, Stdout: true, Stderr: true,
	}, scheme.ParameterCodec)
	executor, err := remotecommand.NewSPDYExecutor(cluster.RESTConfig(), http.MethodPost, request.URL())
	if err != nil {
		return "", fmt.Errorf("prepare exec in %s/%s: %w", namespace, pod, err)
	}
	var stdout, stderr bytes.Buffer
	err = executor.StreamWithContext(ctx, remotecommand.StreamOptions{Stdout: &stdout, Stderr: &stderr})
	if err != nil {
		return stdout.String(), fmt.Errorf("exec in %s/%s: %w; stderr: %s", namespace, pod, err, stderr.String())
	}
	return stdout.String(), nil
}
