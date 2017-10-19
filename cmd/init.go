package cmd

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/Clever/microplane/initialize"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "init short description",
	Long: `init
                long
                description`,
	Run: func(cmd *cobra.Command, args []string) {
		target := "target-" + strconv.Itoa(int(time.Now().Unix()))

		// TODO: Remove this example query
		query := "org:Clever"
		repos, err := initialize.GithubSearch(query)
		if err != nil {
			log.Fatal(err)
		}

		targetDir := "./" + target + "/"
		err = os.Mkdir(targetDir, 0755)
		if err != nil {
			log.Fatal(err)
		}

		path := targetDir + "init.json"
		err = initialize.WriteInitJSON(initialize.Output{
			Repos: repos,
		}, path)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(target)
	},
}
