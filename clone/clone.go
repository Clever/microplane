package clone

import (
	"context"
	"os/exec"
	"path"
)

type Input struct {
	// WorkDir is where results will be stored:
	//   - {WorkDir}/cloned: stores the result of `git clone`
	WorkDir string
	// GitURL to clone.
	GitURL string
}

type Output struct {
	Success       bool
	ClonedIntoDir string
}

type Error struct {
	error
	GitCloneOutput string
}

func Clone(ctx context.Context, input Input) (Output, error) {
	cloneIntoDir := path.Join(input.WorkDir, "cloned")
	cmd := exec.CommandContext(ctx, "git", "clone", input.GitURL, cloneIntoDir)
	cmd.Dir = input.WorkDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return Output{Success: false}, Error{error: err, GitCloneOutput: string(output)}
	}
	return Output{Success: true, ClonedIntoDir: cloneIntoDir}, nil
}
