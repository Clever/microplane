package cmd

import (
	"fmt"
	"log"
	"path"

	"github.com/Clever/microplane/initialize"

	"github.com/spf13/cobra"
)

func initOutputPath() string {
	return path.Join(workDir, "init.json")
}

var initCmd = &cobra.Command{
	Use:   "init [query]",
	Short: "Initialize a microplane workflow",
	Long: `Initialize a microplane workflow. It targets repos based on a Github Search query. For example

$ mp init "org:Clever path:circle.yml"

would target all Clever repos with a circle.yml file.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := args[0]
		output, err := initialize.Initialize(initialize.Input{
			Query:   query,
			WorkDir: workDir,
		})
		if err != nil {
			log.Fatal(err)
		}

		err = writeJSON(output, initOutputPath())
		if err != nil {
			log.Fatal(err)
		}

		for _, repo := range output.Repos {
			fmt.Println(repo.Name)
		}
	},
}
