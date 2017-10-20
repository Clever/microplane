package push

import (
	"context"
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
	// PRHeadBranch is the branch the new commit lives on
	PRHeadBranch string
	// PRBaseBranch is the branch we intend to merge our change into
	PRBaseBranch string
}

// Output from Push()
type Output struct {
	Success        bool
	CommitSHA      string
	PullRequestURL string
}

// Error and details from Push()
type Error struct {
	error
	Details string
}

// Push pushes the commit to Github and opens a pull request
func Push(ctx context.Context, input Input) (Output, error) {
	// Get the commit SHA from the last commit
	cmd := Command{Path: "git", Args: []string{"log", "-1", "--pretty=format:%H"}}
	gitLog := exec.CommandContext(ctx, cmd.Path, cmd.Args...)
	gitLog.Dir = input.PlanDir
	gitLogOutput, err := gitLog.CombinedOutput()
	if err != nil {
		return Output{Success: false}, Error{error: err, Details: string(gitLogOutput)}
	}

	// Push the commit
	cmd = Command{Path: "git", Args: []string{"push"}}
	gitPush := exec.CommandContext(ctx, cmd.Path, cmd.Args...)
	gitPush.Dir = input.PlanDir
	if output, err := gitPush.CombinedOutput(); err != nil {
		return Output{Success: false}, Error{error: err, Details: string(output)}
	}

	// Open a pull request
	cmd = Command{Path: "hub", Args: []string{"pull-request", "-m", input.PRMessage, "-a", input.PRAssignee, "-b", input.PRBaseBranch, "-h", input.PRHeadBranch}}
	hubPullRequest := exec.CommandContext(ctx, cmd.Path, cmd.Args...)
	hubPullRequest.Dir = input.PlanDir
	hubPullRequestOutput, err := hubPullRequest.CombinedOutput()
	if err != nil {
		return Output{Success: false}, Error{error: err, Details: string(hubPullRequestOutput)}
	}

	return Output{Success: true, CommitSHA: string(gitLogOutput), PullRequestURL: strings.TrimSpace(string(hubPullRequestOutput))}, nil
}
