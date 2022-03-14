package main

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
    Users []User
    Organizations []Organization
    Repositories []Repository
}

func main() {}
