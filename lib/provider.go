package lib

import (
	"context"
	"fmt"
	"os"

	"github.com/google/go-github/github"
	"github.com/xanzy/go-gitlab"
	"golang.org/x/oauth2"
)

// ProviderConfig contains the essential parameters that define a provider
type ProviderConfig struct {
	Backend    string
	BackendURL string
}

func (pc ProviderConfig) IsEnterprise() bool {
	return pc.BackendURL != ""
}

func (pc ProviderConfig) CloneURLPrefix() string {
	return fmt.Sprintf("git@%s.com", pc.Backend)
}

// Provider is an abstraction over a Git provider (Github, Gitlab, etc)
type Provider struct {
	ProviderConfig
}

func NewProviderFromConfig(pc ProviderConfig) *Provider {
	return &Provider{
		ProviderConfig: pc,
	}
}

func (p *Provider) GithubClient(ctx context.Context) (*github.Client, error) {
	// validation
	if p.Backend != "github" {
		return nil, fmt.Errorf("cannot initialize GithubClient: backend is not 'github', but instead is '%s'", p.Backend)
	}
	token := os.Getenv("GITHUB_API_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("cannot initialize GithubClient: GITHUB_API_TOKEN is not set")
	}

	// create the client
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	return client, nil
}

func (p *Provider) GitlabClient() (*gitlab.Client, error) {
	// validation
	if p.Backend != "gitlab" {
		return nil, fmt.Errorf("cannot initialize GitlabClient: backend is not 'gitlab', but instead is '%s'", p.Backend)
	}
	token := os.Getenv("GITLAB_API_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("cannot initialize GitlabClient: GITLAB_API_TOKEN is not set")
	}

	// create client
	clientOptions := []gitlab.ClientOptionFunc{}
	if p.IsEnterprise() {
		clientOptions = append(clientOptions, gitlab.WithBaseURL(p.BackendURL))
	}

	return gitlab.NewClient(token, clientOptions...)
}
