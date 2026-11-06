package e2e

import (
	"context"
	"fmt"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	baokubernetes "github.com/openbao/openbao/api/auth/kubernetes/v2"
	bao "github.com/openbao/openbao/api/v2"
	"github.com/stretchr/testify/require"
	authenticationv1 "k8s.io/api/authentication/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

var _ = Describe("Secrets", Label("secrets"), func() {
	const owner = testNamespace
	const path = "test/e2e"

	It("allows the owner namespace to read its secret", func(ctx SpecContext) {
		ownerClient, err := lab.OpenBao()
		Expect(err).NotTo(HaveOccurred())
		Expect(loginAsServiceAccount(ctx, ownerClient, owner, "default")).To(Succeed())

		secret, err := ownerClient.KVv2("secret").Get(ctx, path)
		Expect(err).NotTo(HaveOccurred())
		Expect(secret.Data["value"]).To(Equal("homelab-secret-injection"))
	}, SpecTimeout(5*time.Minute))

	It("rejects writes from the owner namespace", func(ctx SpecContext) {
		ownerClient, err := lab.OpenBao()
		Expect(err).NotTo(HaveOccurred())
		Expect(loginAsServiceAccount(ctx, ownerClient, owner, "default")).To(Succeed())

		_, err = ownerClient.KVv2("secret").Put(ctx, path, map[string]any{"value": "unauthorized-write"})
		var response *bao.ResponseError
		require.ErrorAs(GinkgoTB(), err, &response, "OpenBao must reject this identity with a typed API error")
		Expect(response.StatusCode).To(Equal(http.StatusForbidden))

		secret, err := ownerClient.KVv2("secret").Get(ctx, path)
		Expect(err).NotTo(HaveOccurred())
		Expect(secret.Data["value"]).To(Equal("homelab-secret-injection"))
	}, SpecTimeout(5*time.Minute))

	It("rejects a non-default service account", func(ctx SpecContext) {
		candidate, err := lab.OpenBao()
		Expect(err).NotTo(HaveOccurred())
		err = loginAsServiceAccount(ctx, candidate, owner, "untrusted")
		var response *bao.ResponseError
		require.ErrorAs(GinkgoTB(), err, &response, "OpenBao must reject this identity with a typed API error")
		Expect(response.StatusCode).To(Equal(http.StatusForbidden))
	}, SpecTimeout(5*time.Minute))

	It("injects the owner's secret into a Kubernetes Secret", func(ctx SpecContext) {
		injected, err := lab.Client.CoreV1().Secrets(owner).Create(
			ctx,
			secretReference(path),
			metav1.CreateOptions{},
		)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(deleteInjectedSecret, owner, injected, NodeTimeout(time.Minute))
		Expect(string(injected.Data["value"])).To(Equal("homelab-secret-injection"))
	}, SpecTimeout(5*time.Minute))

	It("preserves ordinary Kubernetes Secrets", func(ctx SpecContext) {
		secret, err := lab.Client.CoreV1().Secrets(owner).Create(ctx, &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{GenerateName: "homelab-test-ordinary-"},
			StringData: map[string]string{"value": "ordinary-secret"},
		}, metav1.CreateOptions{})
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(deleteInjectedSecret, owner, secret, NodeTimeout(time.Minute))
		Expect(string(secret.Data["value"])).To(Equal("ordinary-secret"))
	}, SpecTimeout(5*time.Minute))

	Context("from another namespace", func() {
		const other = otherNamespace

		It("rejects a read of the owner's secret", func(ctx SpecContext) {
			otherClient, err := lab.OpenBao()
			Expect(err).NotTo(HaveOccurred())
			Expect(loginAsServiceAccount(ctx, otherClient, other, "default")).To(Succeed())

			_, err = otherClient.KVv2("secret").Get(ctx, path)
			var response *bao.ResponseError
			require.ErrorAs(GinkgoTB(), err, &response, "OpenBao must reject this identity with a typed API error")
			Expect(response.StatusCode).To(Equal(http.StatusForbidden))
		}, SpecTimeout(5*time.Minute))

		It("rejects injection of the owner's secret", func(ctx SpecContext) {
			injected, err := lab.Client.CoreV1().Secrets(other).Create(
				ctx,
				secretReference(path),
				metav1.CreateOptions{},
			)
			if err == nil {
				DeferCleanup(deleteInjectedSecret, other, injected, NodeTimeout(time.Minute))
			}
			Expect(err).To(MatchError(ContainSubstring("permission denied")))
		}, SpecTimeout(5*time.Minute))
	})
})

func loginAsServiceAccount(ctx context.Context, client *bao.Client, namespace, account string) error {
	jwt, err := lab.Client.CoreV1().ServiceAccounts(namespace).CreateToken(ctx, account,
		&authenticationv1.TokenRequest{
			Spec: authenticationv1.TokenRequestSpec{
				Audiences:         []string{"https://kubernetes.default.svc"},
				ExpirationSeconds: ptr.To(int64(600)),
			},
		}, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("create service-account token: %w", err)
	}
	auth, err := baokubernetes.NewKubernetesAuth("apps", baokubernetes.WithServiceAccountToken(jwt.Status.Token))
	if err != nil {
		return fmt.Errorf("configure Kubernetes auth: %w", err)
	}
	_, err = client.Auth().Login(ctx, auth)

	return err
}

func secretReference(path string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "homelab-test-injection-",
		},
		StringData: map[string]string{"value": "bao:secret/data/" + path + "#value"},
	}
}

func deleteInjectedSecret(ctx context.Context, namespace string, secret *corev1.Secret) error {
	return lab.Client.CoreV1().Secrets(namespace).Delete(ctx, secret.Name, metav1.DeleteOptions{
		Preconditions: &metav1.Preconditions{UID: &secret.UID},
	})
}
