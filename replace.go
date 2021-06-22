package main

import (
	// "fmt"
	"os"
	"path/filepath"
	"strings"
	"io/ioutil"
	"log"
)



func repl(file string) {

	if strings.Contains(file, ".git") {
		return
	}
	if file == "rep.txt" {
		return
	}
	data, err := ioutil.ReadFile(file)
	if err != nil {
		// fmt.Println("File reading error", err)
		return
	}
	dat, err1 := ioutil.ReadFile("rep1.txt")
	if err1 != nil {
		// fmt.Println("File reading error", err1)
		return
	}
	da, err2 := ioutil.ReadFile("with.txt")
	if err2 != nil {
		// fmt.Println("File reading error", err2)
		return
	}
	with := string(da)
	rep := string(dat)
	in := string(data)
	in = strings.Replace(in, rep, with, -1)
	b := []byte(in)
	err3 := ioutil.WriteFile(file, b, 0644)
	if err3 != nil {
		log.Fatal(err3)
	}

}
func main(){
	var files []string

    root := "./mp/"
    err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
        files = append(files, path)
        return nil
	})
    if err != nil {
        panic(err)
    }
    for _, file := range files {
        // fmt.Println(file)
		repl(file)
    }
}



// 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
