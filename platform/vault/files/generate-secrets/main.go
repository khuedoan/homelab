package main

// TODO WIP

// TODO env vars
// export VAULT_ADDR='https://127.0.0.1:8200'
// export VAULT_TOKEN=root

// TODO ACL policy
// path "secret/*" {
//   capabilities = [
//     "create",
//     "list"
//   ]
// }

// TODO config syntax with yaml
// randomPasswords:
//   - path: gitea/admin-password
//     length: 32
//     special: false
//     state: present

import (
	"log"
	// "crypto/rand"

	vault "github.com/hashicorp/vault/api"
	"github.com/sethvargo/go-password/password"
)

func main() {
	config := vault.DefaultConfig()

	config.Address = "http://127.0.0.1:8200"

	client, err := vault.NewClient(config)
	if err != nil {
		log.Fatalf("unable to initialize Vault client: %v", err)
	}

	client.SetToken("root")

	path := "secret/data/gitea/admin-password"

	secret, _ := client.Logical().Read(path)

	if secret == nil {
		res, err := password.Generate(32, 24, 8, false, true)
		if err != nil {
			log.Fatal(err)
		}

		secretData := map[string]interface{}{
			"data": map[string]interface{}{
				"value": res,
			},
		}

		_, err = client.Logical().Write(path, secretData)
		if err != nil {
			log.Fatalf("Unable to write secret: %v", err)
		} else {
			log.Println("Secret written successfully.")
		}
	} else {
		log.Println("Secret already existed.")
	}
}
