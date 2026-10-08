package secrets

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	bao "github.com/openbao/openbao/api/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	coreclient "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func openBaoKV(ctx context.Context, kubeconfig string) (*bao.KVv2, error) {
	if kubeconfig == "" {
		return nil, fmt.Errorf("an explicit kubeconfig path is required")
	}
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("load kubeconfig: %w", err)
	}
	client, err := coreclient.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	secret, err := client.Secrets("openbao").Get(ctx, "openbao-unseal", metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("read OpenBao bootstrap credentials: %w", err)
	}
	token := string(secret.Data["vault-root"])
	if token == "" {
		return nil, fmt.Errorf("OpenBao bootstrap secret has no vault-root key")
	}
	httpClient, err := rest.HTTPClientFor(config)
	if err != nil {
		return nil, err
	}
	baoConfig := bao.NewConfig()
	baoConfig.Address = client.RESTClient().Get().Namespace("openbao").Resource("services").Name("http:openbao:8200").SubResource("proxy").URL().String()
	baoConfig.HttpClient = httpClient
	baoClient, err := bao.NewClient(baoConfig)
	if err != nil {
		return nil, fmt.Errorf("create OpenBao client failed")
	}
	baoClient.SetToken(token)
	return baoClient.KVv2("secret"), nil
}

// Read returns a KV record, or nil when the record does not exist.
func Read(ctx context.Context, kubeconfig, path string) (map[string]string, error) {
	kv, err := openBaoKV(ctx, kubeconfig)
	if err != nil {
		return nil, err
	}
	return readRecord(ctx, kv, path)
}

func readRecord(ctx context.Context, kv *bao.KVv2, path string) (map[string]string, error) {
	record, err := kv.Get(ctx, path)
	if errors.Is(err, bao.ErrSecretNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read OpenBao path %s failed", path)
	}
	data := make(map[string]string, len(record.Data))
	for key, value := range record.Data {
		text, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("OpenBao path %s contains a non-string field", path)
		}
		data[key] = text
	}
	return data, nil
}

// Sync updates OpenBao through the Kubernetes service proxy, preserving unchanged records.
func Sync(ctx context.Context, kubeconfig string, records map[string]map[string]string) error {
	kv, err := openBaoKV(ctx, kubeconfig)
	if err != nil {
		return err
	}
	for path, data := range records {
		existing, err := readRecord(ctx, kv, path)
		if err != nil {
			return err
		}
		if reflect.DeepEqual(existing, data) {
			continue
		}
		fields := make(map[string]any, len(data))
		for key, value := range data {
			fields[key] = value
		}
		if _, err := kv.Put(ctx, path, fields); err != nil {
			return fmt.Errorf("write OpenBao path %s failed", path)
		}
	}
	return nil
}
