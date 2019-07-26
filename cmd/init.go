package cmd

import (
	"fmt"
	"log"

	"github.com/Clever/microplane/initialize"

	"github.com/spf13/cobra"
)

var repoProviderFlag string

var initCmd = &cobra.Command{
	Use:   "init [query]",
	Short: "Initialize a microplane workflow",
	Long: `Initialize a microplane workflow.

## GitHub

It targets repos based on a Github Code Search query. For example

$ mp init "org:Clever filename:circle.yml"

would target all Clever repos with a circle.yml file.

See https://help.github.com/articles/searching-code/ for more details about the search syntax on Github.

## GitLab

If you are using the public version of GitLab, search is done via the Global "projects" scope.
See https://docs.gitlab.com/ce/api/search.html#scope-projects for more information on the search syntax. For example

$ mp init "mp-test-1"

would target a specific repo called mp-test-1.

If you are using an enterprise GitLab instance, we assume you have an ElasticSearch setup.
See https://docs.gitlab.com/ee/user/search/advanced_search_syntax.html for more details about the search syntax on Gitlab.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := args[0]
		output, err := initialize.Initialize(initialize.Input{
			Query:        query,
			WorkDir:      workDir,
			Version:      cliVersion,
			RepoProvider: repoProviderFlag,
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
