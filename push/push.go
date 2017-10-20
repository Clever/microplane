package plan

import (
	"context"
	"os/exec"
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
	// PRMessage is the text of the PR submitted to Github
	PRMessage string
	// PRAssignee is the user who will be assigned the PR
	PRAssignee string
	// PRBaseBranch
	PRBaseBranch string
	// PRHeadBranch
	PRHeadBranch string
}

// Output from Push()
type Output struct {
	Success        bool
	PullRequestURL string
}

// Error and details from Push()
type Error struct {
	error
	Details string
}

// Push pushes the commit to Github and opens a pull request
func Push(ctx context.Context, input Input) (Output, error) {
	// git push origin <branch>
	cmd := Command{Path: "git", Args: []string{"push", "origin", input.PRBaseBranch}}
	gitPush := exec.CommandContext(ctx, cmd.Path, cmd.Args...)
	gitPush.Dir = input.PlanDir
	if output, err := gitPush.CombinedOutput(); err != nil {
		return Output{Success: false}, Error{error: err, Details: string(output)}
	}

	// hub pull-request [-foc] [-b <BASE>] [-h <HEAD>] [-r <REVIEWERS> ] [-a <ASSIGNEES>] [-M <MILESTONE>] [-l <LABELS>]
	cmd = Command{Path: "hub", Args: []string{"pull-request", "-m", input.PRMessage, "-a", input.PRAssignee, "-b", input.PRBaseBranch, "-h", input.PRHeadBranch}}
	hubPullRequest := exec.CommandContext(ctx, cmd.Path, cmd.Args...)
	hubPullRequest.Dir = input.PlanDir
	output, err := hubPullRequest.CombinedOutput()
	if err != nil {
		return Output{Success: false}, Error{error: err, Details: string(output)}
	}

	return Output{Success: true, PullRequestURL: string(output)}, nil
}
