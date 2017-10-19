package initialize

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type Repo struct {
	Name     string
	CloneURL string
}

type Output struct {
	Repos []Repo
}

// GithubSearch queries github and returns a list of matching repos
//
// Search Syntax:
// https://help.github.com/articles/searching-repositories/#search-within-a-users-or-organizations-repositories
// https://help.github.com/articles/understanding-the-search-syntax/
func GithubSearch(query string) ([]Repo, error) {
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

// WriteInitJSON writes the output of the `init` command into a JSON file, for use by later commands
func WriteInitJSON(output Output, path string) error {
	b, err := json.MarshalIndent(output, "", "    ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path, b, 0644)
	if err != nil {
		return err
	}

	return nil
}
