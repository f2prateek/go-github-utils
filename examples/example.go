package main

import (
	"fmt"
	"log"
	"os"

	"github.com/f2prateek/go-github-utils/githubu"
)

func main() {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		log.Fatal("$GITHUB_TOKEN is required")
	}

	done := make(chan struct{}, 1)
	defer close(done)

	g := githubu.WithToken(token)

	c, errc := g.GenerateRepos(done, "segmentio")

	count := 0
	for r := range c {
		count = count + 1
		fmt.Printf("%d\t%s\n", count, *r.Name)
	}
	if err := <-errc; err != nil {
		log.Fatal(err)
	}
}
