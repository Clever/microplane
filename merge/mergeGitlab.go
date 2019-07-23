package merge

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Clever/microplane/push"
	"github.com/xanzy/go-gitlab"
)

// Merge an open MR in Gitlab
// - githubLimiter rate limits the # of calls to Github
// - mergeLimiter rate limits # of merges, to prevent load when submitting builds to CI system
func GitlabMerge(ctx context.Context, input Input, githubLimiter *time.Ticker, mergeLimiter *time.Ticker) (Output, error) {
	// Create Gitlab Client
	ctxFunc := gitlab.WithContext(ctx)

	client := gitlab.NewClient(nil, os.Getenv("GITLAB_API_TOKEN"))
	client.SetBaseURL(os.Getenv("GITLAB_URL"))

	// OK to merge?

	// (1) Check if the PR is mergeable
	<-githubLimiter.C
	pid := fmt.Sprintf("%s/%s", input.Org, input.Repo)
	truePointer := true
	mr, _, err := client.MergeRequests.GetMergeRequest(pid, input.PRNumber, &gitlab.GetMergeRequestsOptions{IncludeDivergedCommitsCount: &truePointer}, ctxFunc)
	if err != nil {
		return Output{Success: false}, err
	}
	if mr.State == "merged" {
		// Success! already merged
		return Output{Success: true, MergeCommitSHA: mr.MergeCommitSHA}, nil
	}

	if mr.MergeStatus != "can_be_merged" {
		return Output{Success: false}, fmt.Errorf("MR is not mergeable")
	}

	// (2) Check commit status
	<-githubLimiter.C
	pipelineStatus, err := push.GetPipelineStatus(client, input.Org, input.Repo, &gitlab.ListProjectPipelinesOptions{SHA: &input.CommitSHA})
	if err != nil || pipelineStatus != "success" {
		return Output{Success: false}, fmt.Errorf("status was not 'success', instead was '%s'", pipelineStatus)
	}

	// // (3) check if PR has been approved by a reviewer
	<-githubLimiter.C
	approvals, _, err := client.MergeRequests.GetMergeRequestApprovals(pid, input.PRNumber, ctxFunc)
	if approvals.ApprovalsRequired == 1 {
		if len(approvals.ApprovedBy) == 0 {
			return Output{Success: false}, fmt.Errorf("PR is not approved. Review state is %s", mr.State)
		}
	}
	// Try to rebase master if Diverged Commits greates that zero
	if mr.DivergedCommitsCount > 0 {
		_, err := client.MergeRequests.RebaseMergeRequest(pid, input.PRNumber, ctxFunc)
		if err != nil {
			return Output{Success: false}, fmt.Errorf("Failed to rebase from master")
		}
	}

	// Merge the PR
	<-mergeLimiter.C
	<-githubLimiter.C
	result, _, err := client.MergeRequests.AcceptMergeRequest(pid, input.PRNumber, &gitlab.AcceptMergeRequestOptions{
		ShouldRemoveSourceBranch: &truePointer,
	}, ctxFunc)
	if err != nil {
		return Output{Success: false}, err
	}

	return Output{Success: true, MergeCommitSHA: result.SHA}, nil
}
