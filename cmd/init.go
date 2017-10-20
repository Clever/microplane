package cmd

import (
	"fmt"
	"log"
	"path"

	"github.com/Clever/microplane/initialize"

	"github.com/spf13/cobra"
)

func initOutputPath(target string) string {
	return path.Join(workDir, target, "init.json")
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "init short description",
	Long: `init
                long
				description`,
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

		err = writeJSON(output, initOutputPath(output.Target))
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(output.Target)
	},
}
