package cmd

import (
	"fmt"
	"log"

	"github.com/Clever/microplane/initialize"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [query]",
	Short: "Initialize a microplane workflow",
	Long: `Initialize a microplane workflow.

There are two ways to init, either (1) from a file or (2) via search

## (1) Init from File

$ mp init -f repos.txt

where repos.txt has lines like:

	clever/repo2
	clever/repo2

## (2) Init via Search

### GitHub Code Search

- Search target repos based on a Github Code Search query.

For example:

$ mp init "org:Clever filename:circle.yml"

would target all Clever repos with a circle.yml file.

See https://help.github.com/articles/searching-code/ for more details about the search syntax on Github.

### Github Repo Search

- Search target repos based on a Github Repo Search query.

To init all repos belonging to a specific org use --all-repos flag.

$ mp init "clever" --all-repos

would target all repos in clever org.

To init repos with additional parameters use --repo-search flag

For example:

$ mp init "org:Clever language:Go" --repo-search

would target all Clever repos written in Go

See https://help.github.com/articles/searching-repositories/ for more details about the search syntax on Github.

### GitLab

Search target repos based on a GitLab search.

If you are using the *public* version of GitLab, search is done via the Global "projects" scope.
See https://docs.gitlab.com/ce/api/search.html#scope-projects for more information on the search syntax. For example

$ mp init "mp-test-1"

would target a specific repo called mp-test-1.

If you are using an *enterprise* GitLab instance, we assume you have an ElasticSearch setup.
See https://docs.gitlab.com/ee/user/search/advanced_search_syntax.html for more details about the search syntax on Gitlab.`,
	Args: cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 && initFlagReposFile == "" {
			log.Fatal("to init via code search (default), you must pass a search query. If init from a repo file, specify a repos file with -f." +
				"If init with a github repo search, include --repo-search flag.")
		}

		if len(args) == 1 && initFlagReposFile != "" {
			log.Fatal("to init via code search (default), you must pass a search query. If init from a repo file, specify a repos file with -f." +
				"If init with a github repo search, include --repo-search flag.")
		}

		query := ""
		if len(args) > 0 {
			query = args[0]
		}

		if initCloneType != "ssh" && initCloneType != "https" {
			log.Fatalf("clone-type must be 'ssh' or 'https' but was %s", initCloneType)
		}

		output, err := initialize.Initialize(initialize.Input{
			AllRepos:      initAllrepos,
			Query:         query,
			WorkDir:       workDir,
			Version:       cliVersion,
			Provider:      initProvider,
			ProviderURL:   initProviderURL,
			ReposFromFile: initFlagReposFile,
			RepoSearch:    initRepoSearch,
			CloneType:     initCloneType,
		})
		if err != nil {
			log.Fatal(err)
		}

		err = writeJSON(output, outputPath("", "init"))
		if err != nil {
			log.Fatal(err)
		}

		for _, repo := range output.Repos {
			fmt.Println(repo.Name)
		}
	},
}

var initFlagReposFile string
var initRepoSearch bool
var initAllrepos bool
var initProvider string
var initProviderURL string
var initCloneType string

func init() {
	initCmd.Flags().StringVarP(&initFlagReposFile, "file", "f", "", "get repos from a file instead of searching")
	initCmd.Flags().BoolVar(&initRepoSearch, "repo-search", false, "get repos from a github repo search")
	initCmd.Flags().BoolVar(&initAllrepos, "all-repos", false, "get all repos for a given org")
	initCmd.Flags().StringVar(&initProvider, "provider", "github", "'github' or 'gitlab'")
	initCmd.Flags().StringVar(&initProviderURL, "provider-url", "", "custom URL for enterprise setups")
	initCmd.Flags().StringVar(&initCloneType, "clone-type", "ssh", "'ssh' or 'https'")
}
