package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/Clever/microplane/clone"
	"github.com/Clever/microplane/lib"
	"github.com/Clever/microplane/merge"
	"github.com/Clever/microplane/plan"
	"github.com/spf13/cobra"
)

var planFlagBranch string
var planFlagDiff bool
var planFlagMessage string
var planFlagParallelism int64

// TODO: Pass these *not* via globals
// these variables are set when the cmd starts running
var (
	branchName    string
	commitMessage string
	changeCmd     string
	changeCmdArgs []string
	isSingleRepo  bool
	showDiff      bool
)

var planCmd = &cobra.Command{
	Use:   "plan [cmd] [args...]",
	Args:  cobra.MinimumNArgs(1),
	Short: "Plan changes by running a command against cloned repos",
	Example: `mp plan -b microplaning -m 'microplane fun' -r app-service -- sh -c /absolute/path/to/script
mp plan -b microplaning -m 'microplane fun' -r app-service -- python /absolute/path/to/script`,
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		var parallelismLimit int64

		changeCmd = args[0]
		if len(args) > 1 {
			changeCmdArgs = args[1:]
		}

		branchName, err = cmd.Flags().GetString("branch")
		if err != nil {
			log.Fatal(err)
		}
		if branchName == "" {
			log.Fatal("--branch is required")
		}

		diff, err := cmd.Flags().GetBool("diff")
		if err != nil {
			log.Fatal(err)
		}
		showDiff = diff

		commitMessage, err = cmd.Flags().GetString("message")
		if err != nil {
			log.Fatal(err)
		}
		if commitMessage == "" {
			log.Fatal("--message is required")
		}

		repos, err := whichRepos(cmd)
		if err != nil {
			log.Fatal(err)
		}
		isSingleRepo = len(repos) == 1

		parallelismLimit, err = cmd.Flags().GetInt64("parallelism")
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("planning %d repos with parallelism limit [%d]", len(repos), parallelismLimit)
		err = parallelizeLimited(repos, planOneRepo, parallelismLimit)
		if err != nil {
			log.Fatalf("%d errors:\n %+v\n", strings.Count(err.Error(), " | ")+1, err)
		}
	},
}

func planOneRepo(r lib.Repo, ctx context.Context) error {
	log.Printf("planning: %s/%s", r.Owner, r.Name)

	// Get previous step's output
	var cloneOutput clone.Output
	if loadJSON(outputPath(r.Name, "clone"), &cloneOutput) != nil || !cloneOutput.Success {
		log.Printf("skipping %s/%s, must successfully clone first", r.Owner, r.Name)
		return nil
	}

	// Exit early if already merged
	var mergeOutput struct {
		merge.Output
		Error string
	}
	if loadJSON(outputPath(r.Name, "merge"), &mergeOutput) == nil && mergeOutput.Success {
		log.Printf("%s/%s - already merged", r.Owner, r.Name)
		return nil
	}

	// Prepare workdir for current step's output
	planOutputPath := outputPath(r.Name, "plan")
	planWorkDir := filepath.Dir(planOutputPath)
	if err := os.MkdirAll(planWorkDir, 0755); err != nil {
		return err
	}

	// Execute
	input := plan.Input{
		RepoName:      r.Name,
		RepoDir:       cloneOutput.ClonedIntoDir,
		WorkDir:       planWorkDir,
		Command:       plan.Command{Path: changeCmd, Args: changeCmdArgs},
		CommitMessage: commitMessage,
		BranchName:    branchName,
	}
	output, err := plan.Plan(ctx, input)
	if err != nil {
		o := struct {
			plan.Output
			Error string
		}{output, err.Error()}
		writeJSON(o, planOutputPath)
		return fmt.Errorf("%s/%s error: %+v", r.Owner, r.Name, err)
	}
	writeJSON(output, planOutputPath)
	if showDiff {
		log.Printf("diffing: %s/%s", r.Owner, r.Name)
		fmt.Println(output.GitDiff)
	}
	return nil
}

func init() {
	planCmd.Flags().StringVarP(&planFlagBranch, "branch", "b", "", "Git branch to commit to")
	planCmd.Flags().BoolVarP(&planFlagDiff, "diff", "d", false, "Show the diffs of the changes made per repo")
	planCmd.Flags().StringVarP(&planFlagMessage, "message", "m", "", "Commit message")
	planCmd.Flags().Int64VarP(&planFlagParallelism, "parallelism", "p", defaultParallelism, "Parallelism limit")
}
