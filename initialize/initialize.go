package initialize

import (
	"bufio"
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
	File    string
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
	repos := []Repo{}
	if input.Query != "" {
		repos, err := githubSearch(input.Query)
		if err != nil {
			return Output{}, err
		}
		sort.Sort(ByName(repos))
	} else if input.File != "" {
		file, err := os.Open(input.File)
		if err != nil {
			return err
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			// Try finding Org and Name
			gitURL := scanner.Text()
			owner, name, err := parseGitUrl(gitURL)
			if err != nil {
				return fmt.Errorf("error parsing %s -- %s", gitURL, err.Error())
			}
			repos = append(repos, Repo{
				CloneURL: gitURL,
				Owner:    owner,
				Name:     name,
			})
		}

		if err := scanner.Err(); err != nil {
			return err
		}
	} else {
		return Output{}, fmt.Errorf("invalid input: neither Query nor File was given")
	}

	return Output{
		Version: input.Version,
		Repos:   repos,
	}, nil
}

func parseGitUrl(url string) (owner string, name string, err error) {
	return "owner", "name", nil
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
	for {
		result, resp, err := client.Search.Code(context.Background(), query, opts)
		if err != nil {
			log.Fatalf("Search.Code returned error: %v", err)
		}
		if result.GetIncompleteResults() {
			log.Fatalf("Github API timed out before completing query")
		}

		for _, codeResult := range result.CodeResults {
			repoCopy := *codeResult.Repository
			allRepos[*codeResult.Repository.Name] = &repoCopy
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
