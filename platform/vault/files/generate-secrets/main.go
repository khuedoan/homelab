package main

import (
	"log"
	// "crypto/rand"

	vault "github.com/hashicorp/vault/api"
)

func main() {
	config := vault.DefaultConfig()

	config.Address = "http://127.0.0.1:8200"

	client, err := vault.NewClient(config)
	if err != nil {
		log.Fatalf("unable to initialize Vault client: %v", err)
	}

	// Authenticate
	// WARNING: This quickstart uses the root token for our Vault dev server.
	// Don't do this in production!
	client.SetToken("root") // TODO use secure token

	secretData := map[string]interface{}{
		"data": map[string]interface{}{
			"value": "verystronkpassword",
		},
	}

	_, err = client.Logical().Write("secret/data/gitea/admin-password", secretData)
	if err != nil {
		log.Fatalf("Unable to write secret: %v", err)
	}
	log.Println("Secret written successfully.")
}
