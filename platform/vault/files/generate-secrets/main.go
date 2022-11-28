package main

import (
	"fmt"
	"log"
	"os"

	vault "github.com/hashicorp/vault/api"
	"github.com/sethvargo/go-password/password"
	"gopkg.in/yaml.v2"
)

type RandomPassword struct {
	Path string
	Data []struct {
		Key    string
		Length int
		Value  string
	}
}

func main() {
	log.Print("Reading config file...")
	data, err := os.ReadFile("./config.yaml")

	if err != nil {
		log.Fatalf("unable to read config file: %v", err)
	}

	randomPasswords := []RandomPassword{}

	err = yaml.Unmarshal([]byte(data), &randomPasswords)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	config := vault.DefaultConfig()

	log.Println("Initializing vault client...")
	client, err := vault.NewClient(config)
	if err != nil {
		log.Fatalf("unable to initialize Vault client: %v", err)
	}

	for _, randomPassword := range randomPasswords {
		path := fmt.Sprintf("/secret/data/%s", randomPassword.Path)

		log.Printf("Checking secret: /secret/%s", randomPassword.Path)
		secret, _ := client.Logical().Read(path)

		if secret == nil {
			log.Println("Creating secret...")

			secretData := map[string]interface{}{
				"data": map[string]interface{}{},
			}
			for _, data := range randomPassword.Data {
				var value = ""
				if data.Value == "" {
					log.Printf("Creating random key %s...", data.Key)
					res, err := password.Generate(data.Length, 3, 3, false, true)
					if err != nil {
						log.Fatal(err)
					}
					value = res
				} else {
					log.Printf("Creating key %s...", data.Key)
					value = data.Value
				}

				secretData["data"].(map[string]interface{})[data.Key] = value
			}

			_, err = client.Logical().Write(path, secretData)

			if err != nil {
				log.Fatalf("Unable to write secret: %v", err)
			} else {
				log.Println("Secret written successfully.")
			}
		} else {
			log.Println("Secret already exists.")
		}
	}
}
