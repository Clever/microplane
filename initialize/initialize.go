package initialize

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"
	"time"

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
	WorkDir       string
	Query         string
	Version       string
	RepoProvider  string
	ReposFromFile string
	RepoSearch    bool
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
	if input.ReposFromFile != "" {
		// Read repos from file
		repos, err = reposFromFile(input)
	} else if input.RepoSearch {
		// Do search with Repo type only
		repos, err = githubRepoSearch(input.Query)
	} else {
		// Do code search
		if input.RepoProvider == "github" {
			repos, err = githubSearch(input.Query)
		} else if input.RepoProvider == "gitlab" {
			repos, err = gitlabSearch(input.Query)
		}
	}

	if err != nil {
		return Output{}, err
	}

	sort.Sort(ByName(repos))
	repos = dedupe(repos)
	return Output{
		Version: input.Version,
		Repos:   repos,
	}, nil
}

func dedupe(repos []Repo) []Repo {
	out := []Repo{}
	seen := map[Repo]struct{}{}

	for _, r := range repos {
		_, isDupe := seen[r]
		if isDupe {
			continue
		}
		seen[r] = struct{}{}
		out = append(out, r)
	}
	return out
}

func reposFromFile(input Input) ([]Repo, error) {
	// read file
	bs, err := ioutil.ReadFile(input.ReposFromFile)
	if err != nil {
		return []Repo{}, err
	}

	repos := []Repo{}
	items := strings.Split(string(bs), "\n")
	for _, item := range items {
		if item == "" {
			// in case file ends with newline, ignore it
			continue
		}
		parts := strings.Split(item, "/")
		if len(parts) != 2 {
			return []Repo{}, fmt.Errorf("unable determine repo from line, expected format '{org}/{repo}': %s", item)
		}
		repos = append(repos, Repo{
			Owner:    parts[0],
			Name:     parts[1],
			CloneURL: fmt.Sprintf("git@%s.com:%s", input.RepoProvider, item),
			Provider: input.RepoProvider,
		})
	}
	return repos, nil
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
		if abuseErr, ok := err.(*github.AbuseRateLimitError); ok {
			var waitTime time.Duration
			if abuseErr.RetryAfter != nil {
				waitTime = *abuseErr.RetryAfter
			} else {
				waitTime = 10 * time.Second
			}
			log.Printf("Triggered Github abuse detection - waiting %v then trying again.\n", waitTime)
			time.Sleep(waitTime)
			continue
		} else if err != nil {
			return []Repo{}, err
		}

		for _, codeResult := range result.CodeResults {
			numProcessedResults = numProcessedResults + 1
			repoCopy := *codeResult.Repository
			allRepos[*codeResult.Repository.Name] = &repoCopy
		}

		incompleteResults := result.GetIncompleteResults()
		if incompleteResults {
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

func githubRepoSearch(query string) ([]Repo, error) {
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
		result, resp, err := client.Search.Repositories(context.Background(), query, opts)
		if abuseErr, ok := err.(*github.AbuseRateLimitError); ok {
			var waitTime time.Duration
			if abuseErr.RetryAfter != nil {
				waitTime = *abuseErr.RetryAfter
			} else {
				waitTime = 10 * time.Second
			}
			log.Printf("Triggered Github abuse detection - waiting %v then trying again.\n", waitTime)
			time.Sleep(waitTime)
			continue
		} else if err != nil {
			return []Repo{}, err
		}

		for _, repoResult := range result.Repositories {
			numProcessedResults = numProcessedResults + 1
			allRepos[*repoResult.Name] = &repoResult
		}

		incompleteResults := result.GetIncompleteResults()
		if incompleteResults {
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
	repoNames := make(map[string]bool)
	opt := &gitlab.SearchOptions{
		PerPage: 20,
		Page:    1,
	}
	if isEnterprise {
		for {
			blobs, resp, err := client.Search.Blobs(query, opt)
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
				if _, ok := repoNames[project.Name]; !ok {
					repos = append(repos, Repo{
						Name:     project.Name,
						Owner:    project.Namespace.FullPath,
						CloneURL: project.SSHURLToRepo,
						Provider: "gitlab",
					})
					repoNames[project.Name] = true
				}
			}
			if resp.CurrentPage >= resp.TotalPages {
				break
			}
			opt.Page = resp.NextPage
		}
	} else {
		for {
			projects, resp, err := client.Search.Projects(query, opt)
			if err != nil {
				fmt.Println(err)
			}
			for _, project := range projects {
				if _, ok := repoNames[project.Name]; !ok {
					repos = append(repos, Repo{
						Name:     project.Name,
						Owner:    project.Namespace.FullPath,
						CloneURL: project.SSHURLToRepo,
						Provider: "gitlab",
					})
					repoNames[project.Name] = true
				}
			}
			if resp.CurrentPage >= resp.TotalPages {
				break
			}
			opt.Page = resp.NextPage
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
