package sync

import (
	"context"
	"time"

	"github.com/Clever/microplane/lib"
	"github.com/Clever/microplane/push"
)

type Output struct {
	CommitSHA                 string
	PullRequestCombinedStatus string
	MergeCommitSHA            string
	Merged                    bool
}

func GithubSyncPush(ctx context.Context, r lib.Repo, po push.Output, repoLimiter *time.Ticker) (Output, error) {
	// Create Github Client
	p := lib.NewProviderFromConfig(r.ProviderConfig)
	client, err := p.GithubClient(ctx)
	if err != nil {
		return Output{}, err
	}

	pr, _, err := client.PullRequests.Get(ctx, r.Owner, r.Name, po.PullRequestNumber)
	if err != nil {
		return Output{}, err
	}

	<-repoLimiter.C
	cs, _, err := client.Repositories.GetCombinedStatus(ctx, r.Owner, r.Name, *pr.Head.SHA, nil)
	if err != nil {
		return Output{}, err
	}

	return Output{
		CommitSHA:                 *pr.Head.SHA,
		PullRequestCombinedStatus: *cs.State,
		MergeCommitSHA:            *pr.MergeCommitSHA,
		Merged:                    *pr.Merged,
	}, nil
}
