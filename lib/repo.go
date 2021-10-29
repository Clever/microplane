package lib

import (
	"fmt"
	"net/url"
	"os"
)

// Repo describes a git Repository with a given Provider
type Repo struct {
	Name     string
	Owner    string
	CloneURL string // consider if we can remove this. ComputedCloneURL is a first step
	ProviderConfig
}

func (r Repo) IsGithub() bool {
	return r.ProviderConfig.Backend == "github"
}

func (r Repo) IsGitlab() bool {
	return r.ProviderConfig.Backend == "gitlab"
}

func (r Repo) ComputedCloneURL(useTokenForClone bool) (string, error) {
	// If we saved a CloneURL retrieved from provider's API, use that
	if r.CloneURL != "" {
		return r.CloneURL, nil
	}

	if useTokenForClone {
		if r.IsGitlab() {
			return "", fmt.Errorf("use-token-for-clone is not valid for gitlab provider")
		}
		token := os.Getenv("GITHUB_API_TOKEN")
		if token == "" {
			return "", fmt.Errorf("cannot generate clone url: GITHUB_API_TOKEN is not set")
		}
		return fmt.Sprintf("https://%s@github.com/%s/%s.git", token, r.Owner, r.Name), nil
	}

	// Otherwise, make our best guess!
	hostname := fmt.Sprintf("%s.com", r.ProviderConfig.Backend)
	if r.ProviderConfig.IsEnterprise() {
		parsed, err := url.Parse(r.ProviderConfig.BackendURL)
		if err != nil {
			return "", err
		}
		hostname = parsed.Hostname()
	}
	return fmt.Sprintf("git@%s:%s/%s", hostname, r.Owner, r.Name), nil
}
