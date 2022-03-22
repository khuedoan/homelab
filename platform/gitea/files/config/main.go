package main

import (
	"fmt"
	"log"
	"os"

	"code.gitea.io/sdk/gitea"
	"gopkg.in/yaml.v2"
)

type User struct {
	Name           string
	TokenSecretRef string
}

type Organization struct {
	Name        string
	Description string
}

type Repository struct {
	Name    string
	Owner   string
	Private bool
	Migrate struct {
		Source string
		Mirror bool
	}
	Webhooks []string
}

type Config struct {
	Users         []User
	Organizations []Organization
	Repositories  []Repository
}

func main() {
	data, err := os.ReadFile("./config.yaml")

	if err != nil {
		log.Fatalf("unable to read config file: %v", err)
	}

	config := Config{}

	err = yaml.Unmarshal([]byte(data), &config)

	if err != nil {
		log.Fatalf("error: %v", err)
	}

	fmt.Println(config)

	// TODO
	url := "https://git.khuedoan.com"
	// url := "http://gitea-http:3000"
	password := "thisisjustfortestingdude"

	options := (gitea.SetBasicAuth("gitea_admin", password))
	client, err := gitea.NewClient(url, options)

	if err != nil {
		log.Fatal(err)
	}

	for _, org := range config.Organizations {
		_, _, err = client.CreateOrg(gitea.CreateOrgOption{
			Name:        org.Name,
			Description: org.Description,
		})

		if err != nil {
			log.Printf("Create organization %s: %s", "testing", err)
		}
	}

	for _, repo := range config.Repositories {
		if repo.Migrate.Source != "" {
			_, _, err = client.MigrateRepo(gitea.MigrateRepoOption{
				RepoName:       repo.Name,
				RepoOwner:      repo.Owner,
				CloneAddr:      repo.Migrate.Source,
				Service:        gitea.GitServicePlain,
				Mirror:         repo.Migrate.Mirror,
				Private:        repo.Private,
				MirrorInterval: "10m",
			})

			if err != nil {
				log.Printf("Migrate %s/%s: %s", repo.Owner, repo.Name, err)
			}
		} else {
			_, _, err = client.AdminCreateRepo(repo.Owner, gitea.CreateRepoOption{
				Name: repo.Name,
				Private: repo.Private,
			})
		}
	}
}
