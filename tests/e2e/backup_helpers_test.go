package e2e

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	velerov1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	velerov2 "github.com/vmware-tanzu/velero/pkg/apis/velero/v2alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/utils/ptr"
	crclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func deleteBackup(ctx context.Context, client crclient.Client, backup *velerov1.Backup) error {
	current := &velerov1.Backup{}
	err := client.Get(ctx, crclient.ObjectKeyFromObject(backup), current)
	if apierrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read test backup for cleanup: %w", err)
	}
	if backup.UID != current.UID {
		return fmt.Errorf("test backup %s was replaced before cleanup", backup.Name)
	}
	if current.Status.StartTimestamp == nil &&
		(current.Status.Phase == "" ||
			current.Status.Phase == velerov1.BackupPhaseNew ||
			current.Status.Phase == velerov1.BackupPhaseQueued) {
		uid, version := backup.UID, current.ResourceVersion
		err := client.Delete(ctx, current, crclient.Preconditions{
			UID:             &uid,
			ResourceVersion: &version,
		})
		if err != nil && !apierrors.IsNotFound(err) {
			return fmt.Errorf("cancel never-started test backup: %w", err)
		}
		return nil
	}
	err = client.Create(ctx, &velerov1.DeleteBackupRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      backup.Name,
			Namespace: "velero",
		},
		Spec: velerov1.DeleteBackupRequestSpec{BackupName: backup.Name},
	})
	if err != nil {
		return fmt.Errorf("request test backup deletion: %w", err)
	}
	err = wait.PollUntilContextCancel(ctx, 2*time.Second, true, func(ctx context.Context) (bool, error) {
		err := client.Get(ctx, crclient.ObjectKeyFromObject(backup), &velerov1.Backup{})
		if apierrors.IsNotFound(err) {
			return true, nil
		}

		return false, err
	})
	if err != nil {
		return fmt.Errorf("wait for test backup deletion: %w", err)
	}
	return nil
}

func waitForBackup(ctx context.Context, client crclient.Client, name string) {
	GinkgoHelper()

	Eventually(func(g Gomega) {
		backup := &velerov1.Backup{}
		err := client.Get(ctx, crclient.ObjectKey{
			Namespace: "velero",
			Name:      name,
		}, backup)
		g.Expect(err).NotTo(HaveOccurred())
		switch backup.Status.Phase {
		case velerov1.BackupPhaseFailed, velerov1.BackupPhasePartiallyFailed, velerov1.BackupPhaseFailedValidation:
			StopTrying(fmt.Sprintf("backup %s failed: %+v", name, backup.Status)).Now()
		}
		g.Expect(backup.Status.Phase).To(
			Equal(velerov1.BackupPhaseCompleted),
			"backup velero/%s: queue position %d; validation errors: %v",
			name,
			backup.Status.QueuePosition,
			backup.Status.ValidationErrors,
		)
	}).
		WithContext(ctx).
		WithTimeout(10 * time.Minute).
		WithPolling(3 * time.Second).
		Should(Succeed())
}

func waitForRestore(ctx context.Context, client crclient.Client, name string) {
	GinkgoHelper()

	Eventually(func(g Gomega) {
		restore := &velerov1.Restore{}
		err := client.Get(ctx, crclient.ObjectKey{
			Namespace: "velero",
			Name:      name,
		}, restore)
		g.Expect(err).NotTo(HaveOccurred())
		switch restore.Status.Phase {
		case velerov1.RestorePhaseFailed, velerov1.RestorePhasePartiallyFailed, velerov1.RestorePhaseFailedValidation:
			StopTrying(fmt.Sprintf("restore %s failed: %+v", name, restore.Status)).Now()
		}
		g.Expect(restore.Status.Phase).To(
			Equal(velerov1.RestorePhaseCompleted),
			"restore velero/%s: validation errors: %v",
			name,
			restore.Status.ValidationErrors,
		)
	}).
		WithContext(ctx).
		WithTimeout(10 * time.Minute).
		WithPolling(3 * time.Second).
		Should(Succeed())
}

func backupClient() crclient.Client {
	GinkgoHelper()
	scheme := runtime.NewScheme()
	Expect(velerov1.AddToScheme(scheme)).To(Succeed())
	Expect(velerov2.AddToScheme(scheme)).To(Succeed())
	client, err := crclient.New(lab.RESTConfig(), crclient.Options{
		Scheme: scheme,
		Log:    GinkgoLogr,
	})
	Expect(err).NotTo(HaveOccurred())

	return client
}

func createBackup(ctx context.Context, client crclient.Client) *velerov1.Backup {
	GinkgoHelper()
	backup := &velerov1.Backup{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "homelab-test-",
			Namespace:    "velero",
		},
		Spec: velerov1.BackupSpec{
			IncludedNamespaces: []string{testNamespace},
			IncludedResources:  []string{"persistentvolumeclaims", "persistentvolumes"},
			LabelSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"homelab.khuedoan.com/test-source": "true"},
			},
			SnapshotMoveData: ptr.To(true),
			TTL:              metav1.Duration{Duration: time.Hour},
		},
	}
	Expect(client.Create(ctx, backup)).To(Succeed())
	DeferCleanup(deleteBackup, client, backup, NodeTimeout(3*time.Minute))

	return backup
}

func createRestore(ctx context.Context, client crclient.Client, backup *velerov1.Backup, restoreNamespace string) *velerov1.Restore {
	GinkgoHelper()
	restore := &velerov1.Restore{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "homelab-test-",
			Namespace:    "velero",
		},
		Spec: velerov1.RestoreSpec{
			BackupName:       backup.Name,
			NamespaceMapping: map[string]string{testNamespace: restoreNamespace},
		},
	}
	Expect(client.Create(ctx, restore)).To(Succeed())
	DeferCleanup(deleteRestore, client, restore, NodeTimeout(time.Minute))

	return restore
}

func deleteRestore(ctx context.Context, client crclient.Client, restore *velerov1.Restore) error {
	err := client.Delete(ctx, restore, crclient.Preconditions{UID: &restore.UID})
	if apierrors.IsNotFound(err) {
		return nil
	}
	return err
}
