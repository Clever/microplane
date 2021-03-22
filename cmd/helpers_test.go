package cmd

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/Clever/microplane/lib"
	"github.com/stretchr/testify/assert"
)

var total = 0
var mutex = sync.RWMutex{}

func doNothingOneRepo(r lib.Repo, ctx context.Context) error {
	// TODO: Write to a channel
	fmt.Println("do nothing: ", r.Name)
	// Side effect ... write a temp file
	mutex.Lock()
	defer mutex.Unlock()
	total++

	return nil
}

func TestParallelize(t *testing.T) {
	repos := []lib.Repo{
		lib.Repo{
			Name: "repo1",
		},
		lib.Repo{
			Name: "repo2",
		},
		lib.Repo{
			Name: "repo3",
		},
	}

	err := parallelize(repos, doNothingOneRepo)
	assert.NoError(t, err)
	assert.Equal(t, len(repos), total)
}
