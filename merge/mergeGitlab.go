package merge

import (
	"context"
	"fmt"
	"time"

	"github.com/Clever/microplane/lib"
	"github.com/Clever/microplane/push"
	gitlab "github.com/xanzy/go-gitlab"
)

// Merge an open MR in Gitlab
// - repoLimiter rate limits the # of calls to Github
// - mergeLimiter rate limits # of merges, to prevent load when submitting builds to CI system
func GitlabMerge(ctx context.Context, input Input, repoLimiter *time.Ticker, mergeLimiter *time.Ticker) (Output, error) {
	// Create client
	p := lib.NewProviderFromConfig(input.Repo.ProviderConfig)
	client, err := p.GitlabClient()
	if err != nil {
		return Output{}, err
	}
	ctxFunc := gitlab.WithContext(ctx)

	// OK to merge?

	// (1) Check if the MR is mergeable
	<-repoLimiter.C
	pid := fmt.Sprintf("%s/%s", input.Repo.Owner, input.Repo.Name)
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
	<-repoLimiter.C
	pipelineStatus, err := push.GetPipelineStatus(client, input.Repo.Owner, input.Repo.Name, &gitlab.ListProjectPipelinesOptions{SHA: &input.CommitSHA})
	if err != nil {
		return Output{Success: false}, err
	}

	if input.RequireBuildSuccess && pipelineStatus != "success" {
		return Output{Success: false}, fmt.Errorf("status was not 'success', instead was '%s'", pipelineStatus)
	}

	// // (3) check if MR has been approved by a reviewer
	<-repoLimiter.C
	approvals, _, err := client.MergeRequests.GetMergeRequestApprovals(pid, input.PRNumber, ctxFunc)
	if err != nil {
		return Output{Success: false}, err
	}

	if input.RequireReviewApproval {
		if approvals.ApprovalsRequired > len(approvals.ApprovedBy) {
			return Output{Success: false}, fmt.Errorf("MR is not approved. Review state is %s", mr.State)
		}
	}
	// Try to rebase master if Diverged Commits greates that zero
	if mr.DivergedCommitsCount > 0 {
		_, err := client.MergeRequests.RebaseMergeRequest(pid, input.PRNumber, ctxFunc)
		if err != nil {
			return Output{Success: false}, fmt.Errorf("Failed to rebase from master")
		}
	}

	// Merge the MR
	<-mergeLimiter.C
	<-repoLimiter.C
	result, _, err := client.MergeRequests.AcceptMergeRequest(pid, input.PRNumber, &gitlab.AcceptMergeRequestOptions{
		ShouldRemoveSourceBranch: &truePointer,
	}, ctxFunc)
	if err != nil {
		return Output{Success: false}, err
	}

	return Output{Success: true, MergeCommitSHA: result.SHA}, nil
}
