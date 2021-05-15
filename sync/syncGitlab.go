package sync

import (
	"context"
	"fmt"
	"time"

	"github.com/Clever/microplane/lib"
	"github.com/Clever/microplane/push"
	"github.com/xanzy/go-gitlab"
)

func GitlabSyncPush(ctx context.Context, r lib.Repo, po push.Output, repoLimiter *time.Ticker) (Output, error) {
	// Create client
	p := lib.NewProviderFromConfig(r.ProviderConfig)
	client, err := p.GitlabClient()
	if err != nil {
		return Output{}, err
	}
	pid := fmt.Sprintf("%s/%s", r.Owner, r.Name)
	mr, _, err := client.MergeRequests.GetMergeRequest(pid, po.PullRequestNumber, nil)
	if err != nil {
		return Output{}, err
	}
	pipelineStatus, err := push.GetPipelineStatus(client, r.Owner, r.Name, &gitlab.ListProjectPipelinesOptions{SHA: &mr.SHA})
	if err != nil {
		return Output{}, err
	}

	return Output{
		CommitSHA:                 mr.SHA,
		PullRequestCombinedStatus: pipelineStatus,
		MergeCommitSHA:            mr.MergeCommitSHA,
		Merged:                    mr.MergeStatus == "merged",
	}, nil
}
