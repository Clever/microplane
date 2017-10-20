package push

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// Command represents a command to run.
type Command struct {
	Path string
	Args []string
}

// Input to Push()
type Input struct {
	// PlanDir is where the git repo that has been modified lives.
	PlanDir string
	// WorkDir is where the work associated with the Push operation happens
	WorkDir string
	// PRMessage is the text of the PR submitted to Github
	PRMessage string
	// PRAssignee is the user who will be assigned the PR
	PRAssignee string
	// RepoOwner is the name of the user who owns the Github repo
	RepoOwner string
	// BranchName is the branch name in Git
	BranchName string
}

// Output from Push()
type Output struct {
	Success        bool
	CommitSHA      string
	PullRequestURL string
}

// Push pushes the commit to Github and opens a pull request
func Push(ctx context.Context, input Input) (Output, error) {
	// Get the commit SHA from the last commit
	cmd := Command{Path: "git", Args: []string{"log", "-1", "--pretty=format:%H"}}
	gitLog := exec.CommandContext(ctx, cmd.Path, cmd.Args...)
	gitLog.Dir = input.PlanDir
	gitLogOutput, err := gitLog.CombinedOutput()
	if err != nil {
		return Output{Success: false}, errors.New(string(gitLogOutput))
	}

	// Push the commit
	gitHeadBranch := fmt.Sprintf("HEAD:%s", input.BranchName)
	cmd = Command{Path: "git", Args: []string{"push", "-f", "origin", gitHeadBranch}}
	gitPush := exec.CommandContext(ctx, cmd.Path, cmd.Args...)
	gitPush.Dir = input.PlanDir
	if output, err := gitPush.CombinedOutput(); err != nil {
		return Output{Success: false}, errors.New(string(output))
	}

	// Open a pull request
	prHeadBranch := fmt.Sprintf("%s:%s", input.RepoOwner, input.BranchName)
	prBaseBranch := fmt.Sprintf("%s:%s", input.RepoOwner, "master")
	cmd = Command{Path: "hub", Args: []string{"pull-request", "-m", input.PRMessage, "-a", input.PRAssignee, "-b", prBaseBranch, "-h", prHeadBranch}}
	hubPullRequest := exec.CommandContext(ctx, cmd.Path, cmd.Args...)
	hubPullRequest.Dir = input.PlanDir
	hubPullRequestOutput, err := hubPullRequest.CombinedOutput()
	if err != nil {
		return Output{Success: false}, errors.New(string(hubPullRequestOutput))
	}

	return Output{Success: true, CommitSHA: string(gitLogOutput), PullRequestURL: strings.TrimSpace(string(hubPullRequestOutput))}, nil
}
