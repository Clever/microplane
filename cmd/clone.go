package cmd

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/Clever/microplane/clone"
	"github.com/Clever/microplane/initialize"
	"github.com/spf13/cobra"
)

func loadJSON(file string, obj interface{}) error {
	bs, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	return json.Unmarshal(bs, obj)
}

var cloneCmd = &cobra.Command{
	Use:   "clone [target]",
	Args:  cobra.ExactArgs(1),
	Short: "clone short description",
	Long: `clone
                long
                description`,
	Run: func(cmd *cobra.Command, args []string) {
		target := args[0]
		targetDir := "./" + target + "/"
		var initOutput initialize.Output
		if err := loadJSON(path.Join(targetDir, "init", "init.json"), &initOutput); err != nil {
			log.Fatal(err)
		}

		var cloneInputs []clone.Input

		singleRepo, err := cmd.Flags().GetString("repo")
		if err != nil {
			valid := false
			for _, r := range initOutput.Repos {
				if r.Name == singleRepo {
					valid = true
				}
			}
			if !valid {
				log.Fatalf("%s not a targeted repo name", singleRepo)
			}
		}

		for _, r := range initOutput.Repos {
			if singleRepo != "" && r.Name != singleRepo {
				continue
			}
			workDir := path.Join(targetDir, r.Name, "clone")
			if err := os.MkdirAll(workDir, 0755); err != nil {
				log.Fatal(err)
			}
			cloneInputs = append(cloneInputs, clone.Input{
				WorkDir: workDir,
				GitURL:  r.CloneURL,
			})
		}

		// TODO: run clone on all inputs
	},
}
