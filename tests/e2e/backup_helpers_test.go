package e2e

import (
	"context"
	"fmt"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/require"
	velerov1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	crclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func deleteBackup(t testing.TB, client crclient.Client, backup *velerov1.Backup) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	current := &velerov1.Backup{}
	err := client.Get(ctx, crclient.ObjectKeyFromObject(backup), current)
	if apierrors.IsNotFound(err) {
		return
	}
	require.NoError(t, err, "read test backup for cleanup")
	require.Equal(t, backup.UID, current.UID, "test backup %s was replaced before cleanup", backup.Name)
	if current.Status.StartTimestamp == nil &&
		(current.Status.Phase == "" || current.Status.Phase == velerov1.BackupPhaseNew || current.Status.Phase == velerov1.BackupPhaseQueued) {
		uid, version := backup.UID, current.ResourceVersion
		err := client.Delete(ctx, current, crclient.Preconditions{UID: &uid, ResourceVersion: &version})
		if !apierrors.IsNotFound(err) {
			require.NoError(t, err, "cancel never-started test backup")
		}
		return
	}
	err = client.Create(ctx, &velerov1.DeleteBackupRequest{
		ObjectMeta: metav1.ObjectMeta{Name: backup.Name, Namespace: "velero"},
		Spec:       velerov1.DeleteBackupRequestSpec{BackupName: backup.Name},
	})
	require.NoError(t, err, "request test backup deletion")
	err = wait.PollUntilContextCancel(ctx, 2*time.Second, true, func(ctx context.Context) (bool, error) {
		err := client.Get(ctx, crclient.ObjectKeyFromObject(backup), &velerov1.Backup{})
		if apierrors.IsNotFound(err) {
			return true, nil
		}
		return false, err
	})
	require.NoError(t, err, "wait for test backup deletion")
}

func waitForBackup(ctx context.Context, client crclient.Client, name string) {
	GinkgoHelper()
	Eventually(func(g Gomega) {
		backup := &velerov1.Backup{}
		err := client.Get(ctx, crclient.ObjectKey{Namespace: "velero", Name: name}, backup)
		g.Expect(err).NotTo(HaveOccurred())
		switch backup.Status.Phase {
		case velerov1.BackupPhaseFailed, velerov1.BackupPhasePartiallyFailed, velerov1.BackupPhaseFailedValidation:
			StopTrying(fmt.Sprintf("backup %s failed: %+v", name, backup.Status)).Now()
		}
		g.Expect(backup.Status.Phase).To(Equal(velerov1.BackupPhaseCompleted), "backup velero/%s: queue position %d; validation errors: %v", name, backup.Status.QueuePosition, backup.Status.ValidationErrors)
	}).WithContext(ctx).WithTimeout(10 * time.Minute).WithPolling(3 * time.Second).Should(Succeed())
}

func waitForRestore(ctx context.Context, client crclient.Client, name string) {
	GinkgoHelper()
	Eventually(func(g Gomega) {
		restore := &velerov1.Restore{}
		err := client.Get(ctx, crclient.ObjectKey{Namespace: "velero", Name: name}, restore)
		g.Expect(err).NotTo(HaveOccurred())
		switch restore.Status.Phase {
		case velerov1.RestorePhaseFailed, velerov1.RestorePhasePartiallyFailed, velerov1.RestorePhaseFailedValidation:
			StopTrying(fmt.Sprintf("restore %s failed: %+v", name, restore.Status)).Now()
		}
		g.Expect(restore.Status.Phase).To(Equal(velerov1.RestorePhaseCompleted), "restore velero/%s: validation errors: %v", name, restore.Status.ValidationErrors)
	}).WithContext(ctx).WithTimeout(10 * time.Minute).WithPolling(3 * time.Second).Should(Succeed())
}
