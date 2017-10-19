package cmd

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path"

	"github.com/spf13/cobra"
)

var workDir string

var RootCmd = &cobra.Command{
	Use:   "microplane",
	Short: "microplane",
	Long: `Microplane
                long
                description`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Microplane RootCmd")
	},
}

func init() {
	RootCmd.PersistentFlags().StringP("repo", "r", "", "single repo to operate on")
	RootCmd.AddCommand(cloneCmd)
	RootCmd.AddCommand(initCmd)
	RootCmd.AddCommand(mergeCmd)
	RootCmd.AddCommand(planCmd)
	RootCmd.AddCommand(pushCmd)
	RootCmd.AddCommand(statusCmd)

	// Determine workDir
	user, err := user.Current()
	if err != nil {
		log.Fatalf("error looking up user: %s\n", err.Error())
	}
	workDir = path.Join(user.HomeDir, "/.microplane")

	// Create workDir, if doesn't yet exist
	if _, err = os.Stat(workDir); os.IsNotExist(err) {
		err = os.Mkdir(workDir, 0755)
		if err != nil {
			log.Fatalf("error creating workDir: %s\n", err.Error())
		}
	}
}

func Execute() error {
	return RootCmd.Execute()
}
