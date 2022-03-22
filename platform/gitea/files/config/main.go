package main

import (
	"log"

	"code.gitea.io/sdk/gitea"
)

type User struct {
	Name           string
	TokenSecretRef string
}

type Organization struct {
	Name    string
	Members []string
}

type Repository struct {
	Path    string
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
	// TODO
	url := "https://git.khuedoan.com"
	// url := "http://gitea-http:3000"
	password := "thisisjustfortestingdude"

	options := (gitea.SetBasicAuth("gitea_admin", password))
	client, err := gitea.NewClient(url, options)

	if err != nil {
		log.Fatal(err)
	}

	_, _, err = client.CreateOrg(gitea.CreateOrgOption{
		Name:        "testing",
		Description: "this is org description",
	})

	if err != nil {
		log.Printf("Create organization %s: %s", "testing", err)
	}

	_, _, err = client.MigrateRepo(gitea.MigrateRepoOption{
		RepoName:       "homelab",
		RepoOwner:      "ops",
		CloneAddr:      "https://github.com/khuedoan/homelab",
		Service:        gitea.GitServicePlain,
		Mirror:         true,
		Private:        false,
		MirrorInterval: "10m",
	})

	if err != nil {
		log.Printf("Migrate %s/%s: %s", "ops", "homelab", err)
	}
}
