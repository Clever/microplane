package initialize

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/google/go-github/github"
	gitlab "github.com/xanzy/go-gitlab"
	"golang.org/x/oauth2"
)

// Repo describes a GithubRepository
type Repo struct {
	Name     string
	Owner    string
	CloneURL string
	Provider string
}

// Input for Initialize
type Input struct {
	WorkDir      string
	Query        string
	Version      string
	RepoProvider string
}

// Output for Initialize
type Output struct {
	Version string
	Repos   []Repo
}

// ByName allows sorting repos by name
type ByName []Repo

func (a ByName) Len() int           { return len(a) }
func (a ByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByName) Less(i, j int) bool { return a[i].Name < a[j].Name }

// Initialize searches Github for matching repos
func Initialize(input Input) (Output, error) {
	var repos []Repo
	var err error
	if input.RepoProvider == "github" {
		repos, err = githubSearch(input.Query)
	} else if input.RepoProvider == "gitlab" {
		repos, err = gitlabSearch(input.Query)
	}
	if err != nil {
		return Output{}, err
	}
	sort.Sort(ByName(repos))
	return Output{
		Version: input.Version,
		Repos:   repos,
	}, nil
}

// githubSearch queries github and returns a list of matching repos
//
// GitHub Code Search Syntax:
// https://help.github.com/articles/searching-code/
func githubSearch(query string) ([]Repo, error) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_API_TOKEN")},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	opts := &github.SearchOptions{}
	allRepos := map[string]*github.Repository{}
	numProcessedResults := 0
	for {
		result, resp, err := client.Search.Code(context.Background(), query, opts)
		if err != nil {
			log.Printf("Search.Code returned error: %v", err)
		}

		for _, codeResult := range result.CodeResults {
			numProcessedResults = numProcessedResults + 1
			repoCopy := *codeResult.Repository
			allRepos[*codeResult.Repository.Name] = &repoCopy
		}

		incompleteResults := result.GetIncompleteResults()
		if incompleteResults {
			log.Println("WARNING: Github API timed out before completing query")
			log.Printf("processed %d of about %d results -- next page is %d", numProcessedResults, *result.Total, resp.NextPage)
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	repos := []Repo{}
	for _, r := range allRepos {
		repos = append(repos, Repo{
			Name:     r.GetName(),
			Owner:    r.Owner.GetLogin(),
			CloneURL: fmt.Sprintf("git@github.com:%s", r.GetFullName()),
			Provider: "github",
		})
	}

	return repos, nil
}

// gitlabSearch queries gitlab and returns a list of matching repos
//
// Gitlab Code Search Syntax:
// https://docs.gitlab.com/ee/user/search/advanced_global_search.html
// https://docs.gitlab.com/ee/user/search/advanced_search_syntax.html
func gitlabSearch(query string) ([]Repo, error) {
	var projectIDs []int

	client := gitlab.NewClient(nil, os.Getenv("GITLAB_API_TOKEN"))
	isEnterprise := false
	if os.Getenv("GITLAB_URL") != "" {
		isEnterprise = true
		client.SetBaseURL(os.Getenv("GITLAB_URL"))
	}

	repos := []Repo{}
	opt := &gitlab.SearchOptions{
		PerPage: 20,
	}
	if isEnterprise {
		blobs, _, err := client.Search.Blobs(query, opt)
		if err != nil {
			fmt.Println(err)
		}
		for _, blob := range blobs {
			if !contains(projectIDs, blob.ProjectID) {
				projectIDs = append(projectIDs, blob.ProjectID)
			}
		}
		for _, i := range projectIDs {
			project, _, err := client.Projects.GetProject(i, nil)
			if err != nil {
				fmt.Println(err)
			}
			repos = append(repos, Repo{
				Name:     project.Name,
				Owner:    project.Namespace.FullPath,
				CloneURL: project.SSHURLToRepo,
				Provider: "gitlab",
			})
		}
	} else {
		projects, _, err := client.Search.Projects(query, opt)
		if err != nil {
			fmt.Println(err)
		}
		for _, project := range projects {
			repos = append(repos, Repo{
				Name:     project.Name,
				Owner:    project.Namespace.FullPath,
				CloneURL: project.SSHURLToRepo,
				Provider: "gitlab",
			})
		}
	}
	return repos, nil
}

func contains(values []int, target int) bool {
	for _, val := range values {
		if val == target {
			return true
		}
	}
	return false
}
