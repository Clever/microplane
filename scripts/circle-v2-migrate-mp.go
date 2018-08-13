package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

func main() {
	repoName := os.Getenv("MICROPLANE_REPO")
	if repoName == "" {
		log.Fatal("expected MICROPLANE_REPO env var to be set")
	}
	fmt.Printf("%s\n", repoName)

	goPath := os.Getenv("GOPATH")

	// undo circle.yml rename from any previous runs of circle-v2-migrate script
	// mv circle.yml.bak circle.yml
	undoCmd := exec.Command("mv", "circle.yml.bak", "circle.yml")
	undoCmd.Run()

	// refresh version of circle-v2-migrate script in repo
	rmScriptCmd := exec.Command("rm", "./circle-v2-migrate")
	rmScriptCmd.Run()

	cpScriptCmd := exec.Command("cp", fmt.Sprintf("%s/src/github.com/Clever/circle-v2-migrate/circle-v2-migrate", goPath), "./")
	cpScriptCmd.Run()

	// run circle-v2-migrate script against repo
	runScriptCmd := exec.Command("./circle-v2-migrate")
	scriptOutput, err := runScriptCmd.CombinedOutput()
	if err != nil {
		log.Fatalf("cannot run script, %s\n", err.Error())
	}
	log.Printf("circle-v2-migrate output: %s\n", string(scriptOutput))

	// remove circle-v2-migrate script from repo
	finalRmScriptCmd := exec.Command("rm", "./circle-v2-migrate")
	finalRmScriptCmd.Run()

	// remove circle.yml.bak from repo
	rmBackupCmd := exec.Command("rm", "./circle.yml.bak")
	rmBackupCmd.Run()

}
