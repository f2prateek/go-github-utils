package main

import (
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/f2prateek/go-github-utils/githubu"
	"github.com/google/go-github/github"
)

func main() {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		log.Fatal("$GITHUB_TOKEN is required")
	}

	done := make(chan struct{}, 2)
	defer close(done)

	g := githubu.WithToken(token)

	filter := regexp.MustCompile("(analytics-)?integration-.*")

	c1, errc1 := g.GenerateRepos(done, "segmentio")
	c2, errc2 := githubu.Filter(done, c1, func(r github.Repository) bool {
		return filter.MatchString(*r.Name)
	})

	count := 0
	for r := range c2 {
		count = count + 1
		fmt.Printf("%d\t%s\n", count, *r.Name)
	}

	if err := <-errc1; err != nil {
		log.Fatal(err)
	}
	if err := <-errc2; err != nil {
		log.Fatal(err)
	}
}
