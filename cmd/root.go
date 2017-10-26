package cmd

import (
	"context"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"

	"github.com/Clever/microplane/initialize"
	"github.com/spf13/cobra"
)

var workDir string
var cliVersion string

var rootCmd = &cobra.Command{
	Use:   "mp",
	Short: "Microplane makes git changes across many repos",
}

func init() {
	if os.Getenv("GITHUB_API_TOKEN") == "" {
		log.Fatalf("GITHUB_API_TOKEN env var is not set. In order to use microplane, create a token (https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/) then set the env var.")
	}

	rootCmd.PersistentFlags().StringP("repo", "r", "", "single repo to operate on")
	rootCmd.AddCommand(cloneCmd)
	rootCmd.AddCommand(docsCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(mergeCmd)

	rootCmd.AddCommand(planCmd)
	planCmd.Flags().StringVarP(&planFlagBranch, "branch", "b", "", "Git branch to commit to")
	planCmd.Flags().StringVarP(&planFlagMessage, "message", "m", "", "Commit message")

	rootCmd.AddCommand(pushCmd)
	pushCmd.Flags().StringVarP(&pushFlagAssignee, "assignee", "a", "", "Github user to assign the PR to")

	rootCmd.AddCommand(statusCmd)

	workDir, _ = filepath.Abs("./mp")

	// Create workDir, if doesn't yet exist
	if _, err := os.Stat(workDir); os.IsNotExist(err) {
		if err := os.Mkdir(workDir, 0755); err != nil {
			log.Fatalf("error creating workDir: %s\n", err.Error())
		}
	}
}

// Execute starts the CLI
func Execute(version string) error {
	cliVersion = version

	// Check if your current workdir was created with an incompatible version of microplane
	var initOutput initialize.Output
	err := loadJSON(initOutputPath(), &initOutput)
	if err != nil {
		// If there's no file, that's OK
		if !os.IsNotExist(err) {
			log.Fatal(err)
		}
	} else {
		if initOutput.Version != cliVersion {
			log.Fatalf("A workdir (%s) exists, created with microplane version %s. This is incompatible with your version %s. Either run again using a compatible version, or remove the workdir and restart.", workDir, initOutput.Version, version)
		}
	}

	// Verify `hub` is installed
	execCmd := exec.CommandContext(context.Background(), "which", "hub")
	_, err = execCmd.CombinedOutput()
	if err != nil {
		log.Fatalf("hub (github.com/github/hub) is required. Please install hub and verify it's in your path")
	}

	// Verify `hub` is configured (at least that it has a config file)
	user, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	if _, err := os.Stat(path.Join(user.HomeDir, ".config/hub")); err != nil {
		log.Fatalf("hub (github.com/github/hub) is not yet configured: ~/.config/hub does not exist. Please create this file with the following contents: \n\n" +
			`github.com:
- user: <your GH user>
  oauth_token: <your token>
  protocol: https`)
	}

	return rootCmd.Execute()
}
