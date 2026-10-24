package e2e

import (
	"testing"

	"github.com/stretchr/testify/require"
	velerov1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	crclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestQueuedBackupCleanupPreservesUnrelatedBackups(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, velerov1.AddToScheme(scheme))
	unrelated := &velerov1.Backup{
		ObjectMeta: metav1.ObjectMeta{Name: "scheduled", Namespace: "velero", UID: "unrelated-backup"},
		Status:     velerov1.BackupStatus{Phase: velerov1.BackupPhaseInProgress},
	}
	backup := &velerov1.Backup{
		ObjectMeta: metav1.ObjectMeta{Name: "owned", Namespace: "velero", UID: "owned-backup"},
		Status:     velerov1.BackupStatus{Phase: velerov1.BackupPhaseQueued},
	}
	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(unrelated, backup).Build()

	deleteBackup(t, client, backup)

	err := client.Get(t.Context(), crclient.ObjectKeyFromObject(backup), &velerov1.Backup{})
	require.True(t, apierrors.IsNotFound(err), "never-started test backup must be cancelled")
	observed := &velerov1.Backup{}
	require.NoError(t, client.Get(t.Context(), crclient.ObjectKeyFromObject(unrelated), observed))
	require.Equal(t, velerov1.BackupPhaseInProgress, observed.Status.Phase)
	require.Equal(t, unrelated.UID, observed.UID)
}
