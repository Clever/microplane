package initialize

import (
	"encoding/json"
	"io/ioutil"
)

type Repo struct {
	Org      string
	Name     string
	CloneURL string
}

type Output struct {
	Repos []Repo
}

func GithubSearch(query string) (Output, error) {
	// Given a github Query URL
	// Get a list of repos
	return Output{}, nil
}

// WriteInitJSON writes the output of the `init` command into a JSON file, for use by later commands
func WriteInitJSON(output Output, path string) error {
	b, err := json.Marshal(output)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path, b, 0644)
	if err != nil {
		return err
	}

	return nil
}
