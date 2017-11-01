package clone

import (
	"context"
	"os"
	"os/exec"
	"path"
)

type Input struct {
	// WorkDir is where results will be stored:
	//   - {WorkDir}/cloned: stores the result of `git clone`
	WorkDir string
	// GitURL to clone.
	GitURL string
	// Force a clone even if already exists (useful for updating in case of changes to master)
	Force bool
}

type Output struct {
	Success       bool
	ClonedIntoDir string
}

type Error struct {
	error
	Details string
}

func Clone(ctx context.Context, input Input) (Output, error) {
	cloneIntoDir := path.Join(input.WorkDir, "cloned")
	if _, err := os.Stat(cloneIntoDir); err == nil {
		// already cloned
		if input.Force {
			if err := os.RemoveAll(cloneIntoDir); err != nil {
				return Output{Success: false}, err
			}
		} else {
			return Output{Success: true, ClonedIntoDir: cloneIntoDir}, nil
		}
	}

	cmd := exec.CommandContext(ctx, "git", "clone", input.GitURL, cloneIntoDir)
	cmd.Dir = input.WorkDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return Output{Success: false}, Error{error: err, Details: string(output)}
	}
	return Output{Success: true, ClonedIntoDir: cloneIntoDir}, nil
}
