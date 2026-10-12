package kanidm

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	coreclient "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/clientcmd"
)

func waitForReady(ctx context.Context, kubeconfig string) error {
	if kubeconfig == "" {
		return fmt.Errorf("an explicit kubeconfig path is required")
	}
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return fmt.Errorf("load kubeconfig: %w", err)
	}
	client, err := coreclient.NewForConfig(config)
	if err != nil {
		return err
	}
	// Argo CD may not have created the namespace or pod yet. Use the caller's
	// deadline for both pod creation and readiness, including in-flight requests.
	err = wait.PollUntilContextCancel(ctx, time.Second, true, func(ctx context.Context) (bool, error) {
		pod, err := client.Pods("kanidm").Get(ctx, "kanidm-0", metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		if pod.DeletionTimestamp != nil {
			return false, nil
		}
		for _, condition := range pod.Status.Conditions {
			if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionTrue {
				return true, nil
			}
		}
		return false, nil
	})
	if err != nil {
		return fmt.Errorf("wait for Kanidm readiness: %w", err)
	}
	return nil
}
