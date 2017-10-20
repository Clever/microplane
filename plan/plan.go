package plan

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
)

// Command represents a command to run.
type Command struct {
	Path string
	Args []string
}

type Input struct {
	// RepoDir is where the git repo to modify lives. It will be copied into WorkDir
	RepoDir string
	// WorkDir is where we will store some results:
	//   - {WorkDir}/plan: stores a copy of repodir but with a new commit containing changes
	WorkDir string
	// Command to run
	Command Command
	// CommitMessage to send to `git commit -m`
	CommitMessage string
}

type Output struct {
	Success bool
	PlanDir string
	GitDiff string
}

type Error struct {
	error
	Details string
}

func Plan(ctx context.Context, input Input) (Output, error) {
	// create a copy of the cloned repo and run all commands there
	// wipe out the directory in case Plan has been run previously
	// but the change command has been edited and you want to run again
	planDir := path.Join(input.WorkDir, "planned")
	if err := os.RemoveAll(planDir); err != nil {
		return Output{Success: false}, Error{error: err, Details: fmt.Sprintf("could not clear directroy %s", planDir)}
	}
	cmd := exec.CommandContext(ctx, "cp", "-r", "./.", planDir) // "./." copies all the contents of the current directory into the target directory
	cmd.Dir = input.RepoDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return Output{Success: false}, Error{error: err, Details: string(output)}
	}

	// run the change command, git add, and git commit
	for _, cmd := range []Command{
		input.Command,
		Command{Path: "git", Args: []string{"add", "-A"}},
		Command{Path: "git", Args: []string{"commit", "-m", input.CommitMessage}},
	} {
		execCmd := exec.CommandContext(ctx, cmd.Path, cmd.Args...)
		execCmd.Dir = planDir
		if output, err := execCmd.CombinedOutput(); err != nil {
			return Output{Success: false}, Error{error: err, Details: string(output)}
		}
	}

	// add the git diff to output, might be useful / convenient?
	var gitDiff string
	gitDiffCmd := exec.CommandContext(ctx, "git", "diff", "HEAD^", "HEAD")
	gitDiffCmd.Dir = planDir
	output, err := gitDiffCmd.CombinedOutput()
	if err != nil {
		return Output{Success: false}, Error{error: err, Details: string(output)}
	}
	gitDiff = string(output)

	return Output{Success: true, PlanDir: planDir, GitDiff: gitDiff}, nil
}
