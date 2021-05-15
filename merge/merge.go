package merge

import (
	"context"
	"fmt"
	"time"

	"github.com/Clever/microplane/lib"
	"github.com/google/go-github/v35/github"
)

// Input to Push()
type Input struct {
	// Repo is the git Repo
	Repo lib.Repo
	// PRNumber of Github, e.g. for https://github.com/Clever/microplane/pull/123, the PRNumber is 123
	PRNumber int
	// CommitSHA for the commit which opened the above PR. Used to look up Commit status.
	CommitSHA string
	// RequireReviewApproval specifies if the PR must be approved before merging
	// - must have at least 1 reviewer
	// - all reviewers must have explicitly approved
	RequireReviewApproval bool
	// RequireBuildSuccess specifies if the PR must have a successful build before merging
	RequireBuildSuccess bool
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
// - repoLimiter rate limits the # of calls to Github
// - mergeLimiter rate limits # of merges, to prevent load when submitting builds to CI system
func GitHubMerge(ctx context.Context, input Input, repoLimiter *time.Ticker, mergeLimiter *time.Ticker) (Output, error) {
	// Create Github Client
	p := lib.NewProviderFromConfig(input.Repo.ProviderConfig)
	client, err := p.GithubClient(ctx)
	if err != nil {
		return Output{}, err
	}

	// OK to merge?

	// (1) Check if the PR is mergeable
	<-repoLimiter.C
	pr, _, err := client.PullRequests.Get(ctx, input.Repo.Owner, input.Repo.Name, input.PRNumber)
	if err != nil {
		return Output{Success: false}, err
	}

	if pr.GetMerged() {
		// Success! already merged
		return Output{Success: true, MergeCommitSHA: pr.GetMergeCommitSHA()}, nil
	}

	if !pr.GetMergeable() {
		return Output{Success: false}, fmt.Errorf("PR is not mergeable")
	}

	// (2) Check commit status
	<-repoLimiter.C
	status, _, err := client.Repositories.GetCombinedStatus(ctx, input.Repo.Owner, input.Repo.Name, input.CommitSHA, &github.ListOptions{})
	if err != nil {
		return Output{Success: false}, err
	}

	if input.RequireBuildSuccess {
		state := status.GetState()
		if state != "success" {
			return Output{Success: false}, fmt.Errorf("Build status was not 'success', instead was '%s'. Use --ignore-build-status to override this check.", state)
		}
	}

	// (3) check if PR has been approved by a reviewer
	<-repoLimiter.C
	reviews, _, err := client.PullRequests.ListReviews(ctx, input.Repo.Owner, input.Repo.Name, input.PRNumber, &github.ListOptions{})
	if input.RequireReviewApproval {
		if len(reviews) == 0 {
			return Output{Success: false}, fmt.Errorf("PR awaiting review. Use --ignore-review-approval to override this check.")
		}
		for _, r := range reviews {
			if r.GetState() != "APPROVED" {
				return Output{Success: false}, fmt.Errorf("PR is not approved. Review state is %s. Use --ignore-review-approval to override this check.", r.GetState())
			}
		}
	}

	// Merge the PR
	options := &github.PullRequestOptions{}
	commitMsg := ""
	<-mergeLimiter.C
	<-repoLimiter.C
	result, _, err := client.PullRequests.Merge(ctx, input.Repo.Owner, input.Repo.Name, input.PRNumber, commitMsg, options)
	if err != nil {
		return Output{Success: false}, err
	}

	if !result.GetMerged() {
		return Output{Success: false}, fmt.Errorf("failed to merge: %s", result.GetMessage())
	}

	// Delete the branch
	<-repoLimiter.C
	_, err = client.Git.DeleteRef(ctx, input.Repo.Owner, input.Repo.Name, "heads/"+*pr.Head.Ref)
	if err != nil {
		return Output{Success: false}, err
	}

	return Output{Success: true, MergeCommitSHA: result.GetSHA()}, nil
}
