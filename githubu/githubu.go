package githubu

import (
	"sync"

	"golang.org/x/oauth2"

	"github.com/google/go-github/github"
)

// Client embeds a github.Client type and exposes utility functions.
type Client struct {
	*github.Client
}

func WithToken(token string) *Client {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(oauth2.NoContext, ts)
	client := github.NewClient(tc)
	return WithClient(client)
}

func WithClient(client *github.Client) *Client {
	return &Client{client}
}

// Emit repos from channel `in` to `out` if `m` returns true.
func Filter(in <-chan github.Repository, in_errc <-chan error,
	m func(r github.Repository) bool) (<-chan github.Repository,
	<-chan error) {
	out := make(chan github.Repository)
	errc := make(chan error, 1)

	go func() {
		var wg sync.WaitGroup

		for r := range in {
			wg.Add(1)
			go func(r github.Repository) {
				defer wg.Done()
				if m(r) {
					out <- r
				}
			}(r)
		}

		errc <- <-in_errc

		go func() {
			wg.Wait()
			close(out)
		}()
	}()

	return out, errc
}

// GenerateRepos finds repositories for `org` and emits it on a channel unless
// signalled to stop on `done`.
func (client *Client) GenerateRepos(done <-chan struct{},
	org string) (<-chan github.Repository, <-chan error) {
	c := make(chan github.Repository)
	errc := make(chan error, 1)

	go func() {
		var wg sync.WaitGroup

		opt := &github.RepositoryListByOrgOptions{
			Type:        "all",
			ListOptions: github.ListOptions{PerPage: 100},
		}

		for {
			select {
			case <-done:
				break
			default:
			}

			newRepos, resp, err := client.Repositories.ListByOrg(org, opt)
			if err != nil {
				errc <- err
				break
			}

			wg.Add(1)
			go func() {
				defer wg.Done()
				for _, repo := range newRepos {
					select {
					case c <- repo:
					case <-done:
						break
					}
				}
			}()

			if resp.NextPage == 0 {
				errc <- nil
				break
			}

			opt.ListOptions.Page = resp.NextPage
		}

		go func() {
			wg.Wait()
			close(c)
		}()
	}()

	return c, errc
}

// GenerateCommits finds commits for `org`/`repo` and emits it on a channel unless
// signalled to stop on `done`.
func (client *Client) GenerateCommits(done <-chan struct{},
	org, repo string) (<-chan github.RepositoryCommit, <-chan error) {
	c := make(chan github.RepositoryCommit)
	errc := make(chan error, 1)

	go func() {
		var wg sync.WaitGroup

		opt := &github.CommitsListOptions{
			ListOptions: github.ListOptions{PerPage: 1},
		}

		for {
			select {
			case <-done:
				break
			default:
			}

			newCommits, resp, err := client.Repositories.ListCommits(org, repo, opt)
			if err != nil {
				errc <- err
				break
			}

			wg.Add(1)
			go func() {
				defer wg.Done()
				for _, commit := range newCommits {
					select {
					case c <- commit:
					case <-done:
						break
					}
				}
			}()

			if resp.NextPage == 0 {
				errc <- nil
				break
			}

			opt.ListOptions.Page = resp.NextPage
		}

		go func() {
			wg.Wait()
			close(c)
		}()
	}()

	return c, errc
}
