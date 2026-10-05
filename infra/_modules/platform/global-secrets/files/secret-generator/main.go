package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"

	"github.com/sethvargo/go-password/password"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const namespace = "global-secrets"

type RandomSecret struct {
	Name string
	Data []struct {
		Key     string
		Length  int
		Special bool
	}
}

func getClient() (*kubernetes.Clientset, error) {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	overrides := &clientcmd.ConfigOverrides{}

	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, overrides).ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("Error building client config: %v", err)
	}

	return kubernetes.NewForConfig(config)
}

func generateRandomPassword(length int, special bool) (string, error) {
	numDigits := int(math.Ceil(float64(length) * 0.2))
	numSymbols := 0

	if special {
		numSymbols = int(math.Ceil(float64(length) * 0.2))
	}

	return password.Generate(length, numDigits, numSymbols, false, true)
}

func readConfigFile(filename string) ([]RandomSecret, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("Unable to read config file: %v", err)
	}

	var randomSecrets []RandomSecret
	err = yaml.Unmarshal(data, &randomSecrets)
	if err != nil {
		return nil, fmt.Errorf("Error parsing config file: %v", err)
	}

	return randomSecrets, nil
}

func createOrUpdateSecret(client *kubernetes.Clientset, name string, randomSecret RandomSecret) error {
	secret, err := client.CoreV1().Secrets(namespace).Get(context.Background(), name, metav1.GetOptions{})

	if err != nil {
		// Secret not found, create a new one
		secretData := map[string][]byte{}

		for _, randomPassword := range randomSecret.Data {
			password, err := generateRandomPassword(randomPassword.Length, randomPassword.Special)
			if err != nil {
				log.Printf("Error generating password for key '%s': %v", randomPassword.Key, err)
				continue
			}

			secretData[randomPassword.Key] = []byte(password)
		}

		newSecret := &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Data: secretData,
		}

		_, err := client.CoreV1().Secrets(namespace).Create(context.Background(), newSecret, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("Unable to create secret: %v", err)
		}
		log.Printf("Secret '%s' created successfully.", name)
	} else {
		// Secret exists, check for new keys
		for _, randomKey := range randomSecret.Data {
			if _, exists := secret.Data[randomKey.Key]; !exists {
				// New key found, generate new password
				log.Printf("New key '%s' found in config for secret '%s', generating new password", randomKey.Key, name)
				password, err := generateRandomPassword(randomKey.Length, randomKey.Special)
				if err != nil {
					log.Printf("Error generating password for key '%s': %v", randomKey.Key, err)
					continue
				}
				secret.Data[randomKey.Key] = []byte(password)
			}
		}

		// Update the secret
		_, err := client.CoreV1().Secrets(namespace).Update(context.Background(), secret, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("Unable to update secret: %v", err)
		}
	}

	return nil
}

func main() {
	configFilename := "./config.yaml"
	randomSecrets, err := readConfigFile(configFilename)
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	client, err := getClient()
	if err != nil {
		log.Fatalf("Unable to create Kubernetes client: %v", err)
	}

	for _, randomSecret := range randomSecrets {
		err := createOrUpdateSecret(client, randomSecret.Name, randomSecret)
		if err != nil {
			log.Printf("Error processing secret %s: %v", randomSecret.Name, err)
		}
	}
}
