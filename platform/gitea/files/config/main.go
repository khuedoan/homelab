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
	Hook bool
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
	webhook_token := os.Getenv("WEBHOOK_TOKEN")

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

		if repo.Hook {
			hooks, _, _ := client.ListRepoHooks(repo.Owner, repo.Name, gitea.ListHooksOptions{})
			if len(hooks) == 0 {
				_, _, err = client.CreateRepoHook(repo.Owner, repo.Name, gitea.CreateHookOption{
					Type: gitea.HookTypeGitea,
					Config: map[string]string{
						"url": "http://el-workflows-listener.tekton-workflows:8080",
						"http_method": "post",
						"content_type": "json",
						"secret": webhook_token,
					},
					Events: []string{
						"create",
						"delete",
						"push",
						"pull_request",
					},
					BranchFilter: "*",
					Active: true,
				})

				if err != nil {
					log.Printf("Create hook %s/%s: %v", repo.Owner, repo.Name, err)
				}
			}
		}
	}
}
