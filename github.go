package git_backup

import (
	"context"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/google/go-github/v43/github"
	"golang.org/x/oauth2"
	"net/url"
	"os"
)

type GithubConfig struct {
	JobName     string `yaml:"job_name"`
	AccessToken string `yaml:"access_token"`
	URL         string `yaml:"url,omitempty"`
	client      *github.Client
}

func (c *GithubConfig) Test() error {
	me, err := c.getMe()
	if err != nil {
		return err
	}
	fmt.Printf("Authenticated with github as: %s", me.Login)
	return nil
}

func (c *GithubConfig) GetName() string {
	return c.JobName
}

func (c *GithubConfig) ListRepositories() ([]*Repository, error) {
	repos, err := c.getAllRepos()
	if err != nil {
		return nil, err
	}
	out := make([]*Repository, len(repos))
	for i, repo := range repos {
		gitUrl, err := url.Parse(*repo.CloneURL)
		if err != nil {
			return out, err
		}
		gitUrl.User = url.UserPassword("github", c.AccessToken)
		out[i] = &Repository{GitURL: *gitUrl}
	}
	return out, nil
}

func (c *GithubConfig) setDefaults() {
	if c.JobName == "" {
		c.JobName = "GitHub"
	}
	httpClient := oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(&oauth2.Token{AccessToken: c.AccessToken}))
	if c.URL == "" {
		c.client = github.NewClient(httpClient)
	} else {
		var err error
		c.client, err = github.NewEnterpriseClient(fmt.Sprintf("%s/api/v3/", c.URL), fmt.Sprintf("%s/api/uploads/", c.URL), httpClient)
		if err != nil {
			panic(err)
		}
	}
}

func (c *GithubConfig) getMe() (*github.User, error) {
	response, _, err := c.client.Users.Get(context.Background(), "")
	return response, err
}

func (c *GithubConfig) getAllRepos() ([]*github.Repository, error) {
	all := make([]*github.Repository, 0)
	var err error

	for repos, response, apiErr := c.getRepos(1); true; repos, response, apiErr = c.getRepos(response.NextPage) {
		if apiErr != nil {
			err = apiErr
			break
		} else {
			all = append(all, repos...)
		}

		if len(repos) == 0 || response.NextPage == 0 {
			break
		}
	}
	if err != nil {
		return all, err
	}

	for repos, response, apiErr := c.getStarredRepos(1); true; repos, response, apiErr = c.getStarredRepos(response.NextPage) {
		if apiErr != nil {
			err = apiErr
			break
		} else {
			all = append(all, repos...)
		}

		if len(repos) == 0 || response.NextPage == 0 {
			break
		}
	}
	return all, err
}

func (c *GithubConfig) getRepos(page int) ([]*github.Repository, *github.Response, error) {
	return c.client.Repositories.List(context.Background(), "", &github.RepositoryListOptions{
		ListOptions: github.ListOptions{
			Page:    page,
			PerPage: 100,
		},
	})
}

func (c *GithubConfig) getStarredRepos(page int) ([]*github.Repository, *github.Response, error) {
	starred, response, err := c.client.Activity.ListStarred(context.Background(), "", &github.ActivityListStarredOptions{
		ListOptions: github.ListOptions{
			Page:    page,
			PerPage: 100,
		},
	})
	if err != nil {
		return nil, response, err
	}
	repos := make([]*github.Repository, len(starred))
	for i, _ := range repos {
		repos[i] = starred[i].Repository
	}
	return repos, response, err
}

func (c *GithubConfig) CloneInto(repo *github.Repository, path string) error {
	auth := &http.BasicAuth{
		Username: "git",
		Password: c.AccessToken,
	}
	gitRepo, err := git.PlainClone(path, false, &git.CloneOptions{
		URL:      *repo.CloneURL,
		Auth:     auth,
		Progress: os.Stdout,
	})
	switch err {
	case transport.ErrEmptyRemoteRepository:
		return nil
	default:
		return err
	case git.ErrRepositoryAlreadyExists:
		fallthrough
	case nil:
	}
	if err == git.ErrRepositoryAlreadyExists {
		if gitRepo, err = git.PlainOpen(path); err != nil {
			return err
		} else if w, err := gitRepo.Worktree(); err != nil {
			return err
		} else if err := w.Pull(&git.PullOptions{
			Auth:     auth,
			Progress: os.Stdout,
		}); err == git.NoErrAlreadyUpToDate {
			return nil
		} else if err != nil {
			return err
		}
	}
	err = gitRepo.Fetch(&git.FetchOptions{
		Auth:     auth,
		Progress: os.Stdout,
		Tags:     git.AllTags,
		Force:    true,
	})
	switch err {
	case git.NoErrAlreadyUpToDate:
		return nil
	default:
		return err
	}
}