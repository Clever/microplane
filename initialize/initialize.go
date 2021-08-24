package initialize

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/Clever/microplane/lib"
	"github.com/google/go-github/v35/github"
	gitlab "github.com/xanzy/go-gitlab"
)

// Input for Initialize
type Input struct {
	AllRepos      bool
	WorkDir       string
	Query         string
	Version       string
	Provider      string
	ProviderURL   string
	ReposFromFile string
	RepoSearch    bool
}

// Output for Initialize
type Output struct {
	Version string
	Repos   []lib.Repo
}

// ByName allows sorting repos by name
type ByName []lib.Repo

func (a ByName) Len() int           { return len(a) }
func (a ByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByName) Less(i, j int) bool { return a[i].Name < a[j].Name }

// Initialize searches Provider for matching repos
func Initialize(input Input) (Output, error) {
	p := lib.NewProviderFromConfig(lib.ProviderConfig{
		Backend:    input.Provider,
		BackendURL: input.ProviderURL,
	})

	var repos []lib.Repo
	var err error
	if input.ReposFromFile != "" {
		// Read repos from file
		repos, err = reposFromFile(p, input.ReposFromFile)
	} else if input.RepoSearch {
		// Do search with Repo type only
		repos, err = githubRepoSearch(p, input.Query)
	} else if input.AllRepos {
		// Do search with Repo type only
		repos, err = githubAllRepoSearch(p, input.Query)
	} else {
		// Do code search
		if p.Backend == "github" {
			repos, err = githubSearch(p, input.Query)
		} else if p.Backend == "gitlab" {
			repos, err = gitlabSearch(p, input.Query)
		} else {
			return Output{}, fmt.Errorf("unsupported provider: %s", p.Backend)
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

func dedupe(repos []lib.Repo) []lib.Repo {
	out := []lib.Repo{}
	seen := map[lib.Repo]struct{}{}

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

func reposFromFile(p *lib.Provider, file string) ([]lib.Repo, error) {
	// read file
	bs, err := ioutil.ReadFile(file)
	if err != nil {
		return []lib.Repo{}, err
	}

	repos := []lib.Repo{}
	items := strings.Split(string(bs), "\n")
	for _, item := range items {
		if item == "" {
			// in case file ends with newline, ignore it
			continue
		}
		parts := strings.Split(item, "/")
		if len(parts) != 2 {
			return []lib.Repo{}, fmt.Errorf("unable determine repo from line, expected format '{org}/{repo}': %s", item)
		}
		repos = append(repos, lib.Repo{
			Owner:          parts[0],
			Name:           parts[1],
			ProviderConfig: p.ProviderConfig,
		})
	}
	return repos, nil
}

// githubSearch queries github and returns a list of matching repos
//
// GitHub Code Search Syntax:
// https://help.github.com/articles/searching-code/
func githubSearch(p *lib.Provider, query string) ([]lib.Repo, error) {
	ctx := context.Background()
	client, err := p.GithubClient(ctx)
	if err != nil {
		return []lib.Repo{}, err
	}

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
			return []lib.Repo{}, err
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
	return getFormattedRepos(p, allRepos), nil
}

func githubRepoSearch(p *lib.Provider, query string) ([]lib.Repo, error) {
	ctx := context.Background()
	client, err := p.GithubClient(ctx)
	if err != nil {
		return []lib.Repo{}, err
	}

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
			return []lib.Repo{}, err
		}

		for _, repoResult := range result.Repositories {
			numProcessedResults = numProcessedResults + 1
			// Archived repositories cannot be editted, so they should not be initialized
			if repoResult.Archived != nil && !(*repoResult.Archived) {
				allRepos[*repoResult.Name] = repoResult
			}
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

	return getFormattedRepos(p, allRepos), nil
}

func githubAllRepoSearch(p *lib.Provider, query string) ([]lib.Repo, error) {
	ctx := context.Background()
	client, err := p.GithubClient(ctx)
	if err != nil {
		return []lib.Repo{}, err
	}

	allRepos := map[string]*github.Repository{}
	opts := &github.RepositoryListByOrgOptions{}
	numProcessedResults := 0

	for {
		result, resp, err := client.Repositories.ListByOrg(context.Background(), query, opts)
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
			return []lib.Repo{}, err
		}

		if err != nil {
			return []lib.Repo{}, err
		}

		for _, repoResult := range result {
			numProcessedResults = numProcessedResults + 1
			// Archived repositories cannot be editted, so they should not be initialized
			if repoResult.Archived != nil && !(*repoResult.Archived) {
				allRepos[*repoResult.Name] = repoResult
			}
			allRepos[*repoResult.Name] = repoResult
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return getFormattedRepos(p, allRepos), nil
}

func getFormattedRepos(p *lib.Provider, allRepos map[string]*github.Repository) []lib.Repo {
	formattedRepos := []lib.Repo{}
	for _, r := range allRepos {
		formattedRepos = append(formattedRepos, lib.Repo{
			Name:           r.GetName(),
			Owner:          r.Owner.GetLogin(),
			ProviderConfig: p.ProviderConfig,
		})
	}
	return formattedRepos
}

// gitlabSearch queries gitlab and returns a list of matching repos
//
// Gitlab Code Search Syntax:
// https://docs.gitlab.com/ee/user/search/advanced_global_search.html
// https://docs.gitlab.com/ee/user/search/advanced_search_syntax.html
func gitlabSearch(p *lib.Provider, query string) ([]lib.Repo, error) {
	client, err := p.GitlabClient()
	if err != nil {
		return nil, err
	}

	repos := []lib.Repo{}
	repoNames := make(map[string]bool)
	opt := &gitlab.SearchOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 20,
			Page:    1,
		},
	}
	if p.IsEnterprise() {
		var projectIDs []int
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
				// Archived repositories cannot be editted, so they should not be initialized
				if project.Archived {
					continue
				}
				if _, ok := repoNames[project.Name]; !ok {
					repos = append(repos, lib.Repo{
						Name:           project.Name,
						Owner:          project.Namespace.FullPath,
						CloneURL:       project.SSHURLToRepo,
						ProviderConfig: p.ProviderConfig,
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
				// Archived repositories cannot be editted, so they should not be initialized
				if project.Archived {
					continue
				}
				if _, ok := repoNames[project.Name]; !ok {
					repos = append(repos, lib.Repo{
						Name:           project.Name,
						Owner:          project.Namespace.FullPath,
						CloneURL:       project.SSHURLToRepo,
						ProviderConfig: p.ProviderConfig,
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
