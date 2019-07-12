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
	Long: `Initialize a microplane workflow. It targets repos based on a Github Code Search query. For example

$ mp init "org:Clever filename:circle.yml"

would target all Clever repos with a circle.yml file.

See https://help.github.com/articles/searching-code/ for more details about the syntax.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		provider, err := cmd.Flags().GetString("provider")
		if err != nil {
			log.Fatal(err)
		}
		if provider != "github" && provider != "gitlab" {
			log.Fatal("--provider must be github or gitlab")
		}
		query := args[0]
		output, err := initialize.Initialize(initialize.Input{
			Query:        query,
			WorkDir:      workDir,
			Version:      cliVersion,
			RepoProvider: provider,
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
