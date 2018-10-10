package cmd

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/Clever/microplane/initialize"
	"github.com/spf13/cobra"
)

var workDir string
var cliVersion string

// Github's rate limit for authenticated requests is 5000 QPH = 83.3 QPM = 1.38 QPS = 720ms/query
// We also use a global limiter to prevent concurrent requests, which trigger Github's abuse detection
var githubLimiter = time.NewTicker(720 * time.Millisecond)

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
	mergeCmd.Flags().StringVarP(&mergeFlagThrottle, "throttle", "t", "1ms", "Throttle number of merges, e.g. '30s' means 1 merge per 30 seconds")
	mergeCmd.Flags().BoolVar(&mergeFlagIgnoreReviewApproval, "ignore-review-approval", false, "Ignore whether or not the review has been approved")
	mergeCmd.Flags().BoolVar(&mergeFlagIgnoreBuildStatus, "ignore-build-status", false, "Ignore whether or not builds are passing")

	rootCmd.AddCommand(planCmd)
	planCmd.Flags().StringVarP(&planFlagBranch, "branch", "b", "", "Git branch to commit to")
	planCmd.Flags().StringVarP(&planFlagMessage, "message", "m", "", "Commit message")

	rootCmd.AddCommand(pushCmd)
	pushCmd.Flags().StringVarP(&pushFlagThrottle, "throttle", "t", "1ms", "Throttle number of pushes, e.g. '30s' means 1 push per 30 seconds")
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
	err := loadJSON(outputPath("", "init"), &initOutput)
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

	return rootCmd.Execute()
}

// outputPath helper constructs the output path string for each step
func outputPath(repoName string, step string) string {
	if step == "init" {
		return path.Join(workDir, "init.json")
	}
	return path.Join(workDir, repoName, step, fmt.Sprintf("%s.json", step))
}
