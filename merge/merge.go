package merge

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/eapache/go-resiliency/retrier"
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
	// PRNumber of Github, e.g. for https://github.com/Clever/microplane/pull/123, the PRNumber is 123
	PRNumber int
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

	// Configure a retrier - exponential backoff upon hitting Github API rate limit errors
	r := retrier.New(retrier.ExponentialBackoff(10, time.Second), retrier.WhitelistClassifier{
		&github.RateLimitError{},
		&github.AbuseRateLimitError{},
	})
	r.SetJitter(0.5)

	// OK to merge?

	// (1) Check if the PR is mergeable
	var pr *github.PullRequest
	err := r.Run(func() error {
		var prErr error
		pr, _, prErr = client.PullRequests.Get(ctx, input.Org, input.Repo, input.PRNumber)
		return prErr
	})
	if err != nil {
		return Output{Success: false}, err
	}

	if !pr.GetMergeable() {
		return Output{Success: false}, fmt.Errorf("PR is not mergeable")
	}

	// (2) Check commit status
	var status *github.CombinedStatus
	err = r.Run(func() error {
		var csErr error
		opt := &github.ListOptions{}
		status, _, csErr = client.Repositories.GetCombinedStatus(ctx, input.Org, input.Repo, input.CommitSHA, opt)
		return csErr
	})
	if err != nil {
		return Output{Success: false}, err
	}

	state := status.GetState()
	if state != "success" {
		return Output{Success: false}, fmt.Errorf("status was not 'success', instead was '%s'", state)
	}

	// (3) check if PR has been rejected by a reviewer
	// TODO

	// Merge the PR
	options := &github.PullRequestOptions{}
	commitMsg := ""
	var result *github.PullRequestMergeResult
	err = r.Run(func() error {
		var mergeErr error
		result, _, mergeErr = client.PullRequests.Merge(ctx, input.Org, input.Repo, input.PRNumber, commitMsg, options)
		return mergeErr
	})
	if err != nil {
		return Output{Success: false}, err
	}

	if !result.GetMerged() {
		return Output{Success: false}, fmt.Errorf("failed to merge: %s", result.GetMessage())
	}

	// Delete the branch
	err = r.Run(func() error {
		_, deleteErr := client.Git.DeleteRef(ctx, input.Org, input.Repo, "heads/"+*pr.Head.Ref)
		return deleteErr
	})
	if err != nil {
		return Output{Success: false}, err
	}

	return Output{Success: true, MergeCommitSHA: result.GetSHA()}, nil
}
