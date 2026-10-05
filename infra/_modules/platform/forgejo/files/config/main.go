package main

import (
	"log"
	"net/http"
	"os"

	"codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v3"
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
		log.Fatal(err)
	}

	config := Config{}
	if err := yaml.UnmarshalStrict(data, &config); err != nil {
		log.Fatal(err)
	}

	client, err := forgejo.NewClient(os.Getenv("FORGEJO_HOST"),
		forgejo.SetBasicAuth(os.Getenv("FORGEJO_USER"), os.Getenv("FORGEJO_PASSWORD")))
	if err != nil {
		log.Fatal(err)
	}

	for _, org := range config.Organizations {
		_, response, err := client.GetOrg(org.Name)
		if err == nil {
			continue
		}
		if response == nil || response.StatusCode != http.StatusNotFound {
			log.Fatal(err)
		}
		_, _, err = client.CreateOrg(forgejo.CreateOrgOption{
			Name:        org.Name,
			Description: org.Description,
		})
		if err != nil {
			log.Fatalf("create organization %s: %v", org.Name, err)
		}
	}

	for _, repo := range config.Repositories {
		_, response, err := client.GetRepo(repo.Owner, repo.Name)
		if err == nil {
			continue
		}
		if response == nil || response.StatusCode != http.StatusNotFound {
			log.Fatal(err)
		}
		if repo.Migrate.Source != "" {
			_, _, err = client.MigrateRepo(forgejo.MigrateRepoOption{
				RepoName:       repo.Name,
				RepoOwner:      repo.Owner,
				CloneAddr:      repo.Migrate.Source,
				Service:        forgejo.GitServicePlain,
				Mirror:         repo.Migrate.Mirror,
				Private:        repo.Private,
				MirrorInterval: "10m",
			})
		} else {
			_, _, err = client.AdminCreateRepo(repo.Owner, forgejo.CreateRepoOption{
				Name:    repo.Name,
				Private: repo.Private,
			})
		}
		if err != nil {
			log.Fatalf("create repository %s/%s: %v", repo.Owner, repo.Name, err)
		}
	}
}
