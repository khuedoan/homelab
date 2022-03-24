package main

// TODO WIP clean this up

import (
	"log"
	"os"

	"code.gitea.io/sdk/gitea"
	"gopkg.in/yaml.v2"
)

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
}

type Config struct {
	Organizations []Organization
	Repositories  []Repository
}

func main() {
	data, err := os.ReadFile("./config.yaml")

	if err != nil {
		log.Fatalf("Unable to read config file: %v", err)
	}

	config := Config{}

	err = yaml.Unmarshal([]byte(data), &config)

	if err != nil {
		log.Fatalf("error: %v", err)
	}

	gitea_host := os.Getenv("GITEA_HOST")
	gitea_user := os.Getenv("GITEA_USER")
	gitea_password := os.Getenv("GITEA_PASSWORD")

	options := (gitea.SetBasicAuth(gitea_user, gitea_password))
	client, err := gitea.NewClient(gitea_host, options)

	if err != nil {
		log.Fatal(err)
	}

	for _, org := range config.Organizations {
		_, _, err = client.CreateOrg(gitea.CreateOrgOption{
			Name:        org.Name,
			Description: org.Description,
		})

		if err != nil {
			log.Printf("Create organization %s: %v", org.Name, err)
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
				log.Printf("Migrate %s/%s: %v", repo.Owner, repo.Name, err)
			}
		} else {
			_, _, err = client.AdminCreateRepo(repo.Owner, gitea.CreateRepoOption{
				Name: repo.Name,
				// Description: "TODO",
				Private: repo.Private,
			})
		}
	}
}
