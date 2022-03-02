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

import (
	"fmt"
	"log"

	vault "github.com/hashicorp/vault/api"
	"github.com/sethvargo/go-password/password"
	"gopkg.in/yaml.v2"
)

var data = `
- path: gitea/admin
  key: password
  length: 32
  special: true
- path: gitea/renovate
  key: id
  length: 32
  special: false
- path: gitea/renovate
  key: token
  length: 32
  special: false
`

type RandomPassword struct {
	Path    string `yaml:"path"`
	Length  int    `yaml:"length"`
	Special bool   `yaml:"special"`
}

func main() {
	randomPasswords := []RandomPassword{}

	err := yaml.Unmarshal([]byte(data), &randomPasswords)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	config := vault.DefaultConfig()

	config.Address = "http://127.0.0.1:8200"

	client, err := vault.NewClient(config)
	if err != nil {
		log.Fatalf("unable to initialize Vault client: %v", err)
	}

	client.SetToken("root")

	for _, randomPassword := range randomPasswords {
		path := fmt.Sprintf("/secret/data/%s", randomPassword.Path)

		secret, _ := client.Logical().Read(path)

		if secret == nil {
			res, err := password.Generate(32, 3, 3, false, true)
			if err != nil {
				log.Fatal(err)
			}

			secretData := map[string]interface{}{
				"data": map[string]interface{}{
					"password": res,
				},
			}

			_, err = client.Logical().Write(path, secretData)
			if err != nil {
				log.Fatalf("Unable to write secret: %v", err)
			} else {
				log.Println("Secret written successfully.")
			}
		} else {
			log.Println("Key abc in secret already existed.")
		}
	}
}
