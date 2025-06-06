package ghclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/alesr/ghist/internal/service"
)

const (
	githubAPI  = "https://api.github.com/users/%s/repos"
	ctxTimeout = time.Second * 10
)

type ghAPIError struct {
	Message          string `json:"message"`
	DocumentationURL string `json:"documentation_url"`
	StatusCode       string `json:"status"`
}

func (e ghAPIError) Error() string {
	return fmt.Sprintf("message: '%s', documentation url: '%s', status code: '%s", e.Message, e.DocumentationURL, e.StatusCode)
}

type Client struct{ httpCli *http.Client }

func New(httpCli *http.Client) *Client { return &Client{httpCli: httpCli} }

func (c *Client) FetchRepos(ctx context.Context, username string) ([]service.GithubRepo, error) {
	var repos []service.GithubRepo
	page := 1
	perPage := 100

	for {
		u, err := url.Parse(fmt.Sprintf(githubAPI, username))
		if err != nil {
			return nil, fmt.Errorf("could not parse url for fetching user's repos: %w", err)
		}

		q := u.Query()
		q.Set("page", fmt.Sprintf("%d", page))
		q.Set("per_page", fmt.Sprintf("%d", perPage))
		u.RawQuery = q.Encode()

		ctxReq, cancel := context.WithTimeout(ctx, ctxTimeout)
		defer cancel()

		req, err := http.NewRequestWithContext(ctxReq, http.MethodGet, u.String(), nil)
		if err != nil {
			return nil, fmt.Errorf("could not fetch user's repos: %w", err)
		}

		req.Header.Set("User-Agent", "Ghist")

		resp, err := c.httpCli.Do(req)
		if err != nil {
			return nil, fmt.Errorf("could not get user's repos: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			var e ghAPIError
			if err := json.NewDecoder(resp.Body).Decode(&e); err != nil {
				b, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf("unable to parse API error message '%s' for status '%s': %w", b, resp.Status, err)
				}
			}
			return nil, fmt.Errorf("unexpected status code with error: %w", e)
		}

		var pageRepos []service.GithubRepo
		if err := json.NewDecoder(resp.Body).Decode(&pageRepos); err != nil {
			return nil, fmt.Errorf(("could not decode response body: %w"), err)
		}

		repos = append(repos, pageRepos...)
		if len(pageRepos) < perPage {
			break
		}
		page++
	}
	return repos, nil
}
