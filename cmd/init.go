package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"path"

	"github.com/Clever/microplane/initialize"

	"github.com/spf13/cobra"
)

// writeOutputJSON writes the output of the command into a JSON file, for use by later commands
func writeOutputJSON(output initialize.Output, path string) error {
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

		p := path.Join(workDir, "/", output.Target, "/init.json")
		err = writeOutputJSON(output, p)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(output.Target)
	},
}
