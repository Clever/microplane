package merge

import (
	"context"
	"fmt"
	"os"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// Command represents a command to run.
type Command struct {
	Path string
	Args []string
}

// Input to Push()
type Input struct {
	// Org on Github, e.g. "Clever"
	Org string
	// Repo is the name of the repo on Github, e.g. "microplane"
	Repo string
	// PullRequestNumber of Github, e.g. 1 for https://github.com/Clever/microplane/pull/1
	PullRequestNumber int
	// CommitSHA for the commit which opened the above PR. Used to look up Commit status.
	CommitSHA string
}

// Output from Push()
type Output struct {
	Success        bool
	MergeCommitSHA string
}

// Error and details from Push()
type Error struct {
	error
	Details string
}

// Merge an open PR in Github
func Merge(ctx context.Context, input Input) (Output, error) {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_API_TOKEN")},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	// OK to merge?

	// Check if the PR is mergeable
	pr, _, err := client.PullRequests.Get(ctx, input.Org, input.Repo, input.PullRequestNumber)
	if err != nil {
		return Output{}, err
	}
	if !pr.GetMergeable() {
		return Output{}, fmt.Errorf("PR is not mergeable")
	}

	// Check commit status
	opt := &github.ListOptions{}
	status, _, err := client.Repositories.GetCombinedStatus(ctx, input.Org, input.Repo, input.CommitSHA, opt)
	if err != nil {
		return Output{}, err
	}
	state := status.GetState()
	if state != "success" {
		return Output{}, fmt.Errorf("status was not 'success', instead was '%s'", state)
	}

	// TODO: check if PR has been "rejected" by a reviewer

	///////////////
	// Merge the PR

	// # https://developer.github.com/v3/repos/merging/

	// REPO=$1
	// PR_NUMBER=$2

	// URL="https://api.github.com/repos/Clever/$REPO/pulls/$PR_NUMBER/merge"

	// echo "Merging PR $PR_NUMBER into 'master' for repo Clever/$REPO..."
	// echo "PUT $URL"

	// curl "$URL" \
	//   -XPUT \
	//   -H "Authorization: token $GITHUB_API_TOKEN" \
	//   -d "{
	// 	\"merge_method\": \"rebase\", ?? merge
	// 	\"commit_title\": \"Merge pull request #$PR_NUMBER from Clever/use-golang-1.8\"
	//   }"

	return Output{Success: true, MergeCommitSHA: "TODO"}, nil
}
