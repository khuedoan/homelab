package kanidm

import (
	"context"
	"fmt"
	"net/url"
	"slices"
	"strings"
)

// UserConfig describes a new person and their existing access groups.
type UserConfig struct {
	Username    string
	DisplayName string
	Email       string
	Groups      []string
}

// CreateUser creates a person and returns their one-use credential setup link.
// Existing people are rejected without changing credentials or memberships.
func CreateUser(ctx context.Context, options Options, config UserConfig) (string, error) {
	if !groupNamePattern.MatchString(config.Username) {
		return "", fmt.Errorf("username must start with a lowercase letter and contain at most 64 lowercase letters, digits, dots, underscores, or hyphens")
	}
	if strings.TrimSpace(config.DisplayName) == "" {
		return "", fmt.Errorf("display name is required")
	}
	for _, group := range config.Groups {
		if !groupNamePattern.MatchString(group) {
			return "", fmt.Errorf("invalid access group %q", group)
		}
	}
	c, err := connect(ctx, options)
	if err != nil {
		return "", err
	}
	return c.createUser(ctx, config)
}

func (c *client) createUser(ctx context.Context, config UserConfig) (string, error) {
	var existing *entry
	if err := c.request(ctx, "GET", "/v1/person/"+config.Username, nil, &existing); err != nil {
		return "", err
	}
	if existing != nil {
		return "", fmt.Errorf("user %s already exists; credentials and memberships were left unchanged", config.Username)
	}
	groups := []string{"users"}
	for _, group := range config.Groups {
		if !slices.Contains(groups, group) {
			groups = append(groups, group)
		}
	}
	for _, name := range groups {
		var group *entry
		if err := c.request(ctx, "GET", "/v1/group/"+name, nil, &group); err != nil {
			return "", err
		}
		if group == nil {
			return "", fmt.Errorf("access group %s does not exist; run sso group ensure %s first", name, name)
		}
	}
	person := entry{Attrs: map[string][]string{
		"name": {config.Username}, "displayname": {config.DisplayName}, "mail": {config.Email},
	}}
	if err := c.request(ctx, "POST", "/v1/person", person, nil); err != nil {
		return "", err
	}
	for _, group := range groups {
		if err := c.request(ctx, "POST", "/v1/group/"+group+"/_attr/member", []string{config.Username}, nil); err != nil {
			return "", fmt.Errorf("user %s was created, but adding group %s failed; finish onboarding manually: %w", config.Username, group, err)
		}
	}
	var intent struct {
		Token string `json:"token"`
	}
	if err := c.request(ctx, "GET", "/v1/person/"+config.Username+"/_credential/_update_intent", nil, &intent); err != nil {
		return "", fmt.Errorf("user %s was created, but issuing the credential setup link failed; finish onboarding manually: %w", config.Username, err)
	}
	if intent.Token == "" {
		return "", fmt.Errorf("user %s was created, but Kanidm returned an empty credential setup token; finish onboarding manually", config.Username)
	}
	return c.url + "/ui/reset?" + url.Values{"token": {intent.Token}}.Encode(), nil
}
