package initialize

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// Repo describes a GithubRepository
type Repo struct {
	Name     string
	Owner    string
	CloneURL string
}

// Input for Initialize
type Input struct {
	WorkDir string
	Query   string
	Version string
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
	repos, err := githubSearch(input.Query)
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
		})
	}

	return repos, nil
}
