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

// GenerateRepos finds repositories for `org` and emits it on a channel unless
// signalled to stop on `done`.
func (g *Client) GenerateRepos(done <-chan struct{},
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

			newRepos, resp, err := g.Repositories.ListByOrg(org, opt)
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