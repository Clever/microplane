package lib

// Repo describes a git Repository with a given Provider
type Repo struct {
	Name     string
	Owner    string
	CloneURL string // TODO: someday, compute this from ProviderConfig
	ProviderConfig
}

func (r Repo) IsGithub() bool {
	return r.ProviderConfig.Backend == "github"
}

func (r Repo) IsGitlab() bool {
	return r.ProviderConfig.Backend == "gitlab"
}
