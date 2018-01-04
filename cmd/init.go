package cmd

import (
	"fmt"
	"log"

	"github.com/Clever/microplane/initialize"

	"github.com/spf13/cobra"
)

var initFlagFile string
var initFlagSearch string

var initCmd = &cobra.Command{
	Use:   "init [query]",
	Short: "Initialize a microplane workflow",
	Long: `Initialize a microplane workflow by selecting the repos.

There are two ways to select repos:

(1) Search for them via a Github Code Search query. For example

	$ mp init --search "org:Clever filename:circle.yml"

would target all Clever repos with a circle.yml file.

See https://help.github.com/articles/searching-code/ for more details about the syntax.

(2) pass a file containing repos to target

	$ mp init --file repos.txt

This file should contain list of clone URLs, like so:

	git@github.com:org1/repo-a.git
	git@github.com:org2/repo-b.git
`,
	Args: cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		query := args[0]
		output, err := initialize.Initialize(initialize.Input{
			Query:   query,
			WorkDir: workDir,
			Version: cliVersion,
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
