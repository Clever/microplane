package initialize

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type Repo struct {
	Name     string
	CloneURL string
}

type Input struct {
	WorkDir string
	Query   string
}

type Output struct {
	Target string
	Repos  []Repo
}

func Initialize(input Input) (Output, error) {
	target := "target-" + strconv.Itoa(int(time.Now().Unix()))

	// Create Target dir
	err := os.Mkdir(input.WorkDir+"/"+target+"/", 0755)
	if err != nil {
		return Output{}, err
	}

	repos, err := githubSearch(input.Query)
	if err != nil {
		return Output{}, err
	}

	return Output{
		Target: target,
		Repos:  repos,
	}, nil
}

// githubSearch queries github and returns a list of matching repos
//
// Search Syntax:
// https://help.github.com/articles/searching-repositories/#search-within-a-users-or-organizations-repositories
// https://help.github.com/articles/understanding-the-search-syntax/
func githubSearch(query string) ([]Repo, error) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_API_TOKEN")},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	opts := &github.SearchOptions{}
	allRepos := []github.Repository{}
	for {
		result, resp, err := client.Search.Repositories(context.Background(), query, opts)
		if err != nil {
			log.Fatalf("Search.Repositories returned error: %v", err)
		}
		if result.GetIncompleteResults() {
			log.Fatalf("Github API timed out before completing query")
		}

		allRepos = append(allRepos, result.Repositories...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage

		// TODO: Remove this short-circuiting
		if opts.Page > 2 {
			break
		}

		// TODO: Handle ratelimiting
	}

	repos := []Repo{}
	for _, r := range allRepos {
		repos = append(repos, Repo{
			Name:     r.GetName(),
			CloneURL: r.GetCloneURL(),
		})
	}

	return repos, nil
}
