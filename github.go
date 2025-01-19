package git_backup

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"slices"
	"strings"

	"github.com/google/go-github/v43/github"
	"golang.org/x/oauth2"
)

type GithubConfig struct {
	JobName      string   `yaml:"job_name"`
	AccessToken  string   `yaml:"access_token"`
	URL          string   `yaml:"url,omitempty"`
	Starred      *bool    `yaml:"starred,omitempty"`
	OrgMember    *bool    `yaml:"org_member,omitempty"`
	Collaborator *bool    `yaml:"collaborator,omitempty"`
	Owned        *bool    `yaml:"owned,omitempty"`
	Exclude      []string `yaml:"exclude,omitempty"`
	client       *github.Client
}

func (c *GithubConfig) Test() error {
	me, err := c.getMe()
	if err != nil {
		return err
	}
	log.Printf("Authenticated with github as: %s", *me.Login)
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
	out := make([]*Repository, 0, len(repos))
	for _, repo := range repos {
		gitUrl, err := url.Parse(*repo.CloneURL)
		if err != nil {
			return out, err
		}
		gitUrl.User = url.UserPassword("github", c.AccessToken)

		isExcluded := slices.ContainsFunc(c.Exclude, func(s string) bool {
			if strings.EqualFold(s, *repo.FullName) {
				return true
			}

			if strings.Contains(s, "/") {
				return false
			}

			repoOwner = *repo.FullName[:strings.Index(*repo.FullName, "/")]
			return strings.EqualFold(s, repoOwner)
		})
		if isExcluded {
			log.Printf("Skipping excluded repository: %s", *repo.FullName)
		} else {
			out = append(out, &Repository{
				FullName: *repo.FullName,
				GitURL:   *gitUrl,
			})
		}
	}
	return out, nil
}

func (c *GithubConfig) setDefaults() {
	if c.JobName == "" {
		c.JobName = "GitHub"
	}
	if c.Collaborator == nil {
		c.Collaborator = boolPointer(true)
	}
	if c.OrgMember == nil {
		c.OrgMember = boolPointer(true)
	}
	if c.Owned == nil {
		c.Owned = boolPointer(true)
	}
	if c.Starred == nil {
		c.Starred = boolPointer(true)
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

	if *c.Starred {
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
	}

	return all, err
}

func (c *GithubConfig) getRepos(page int) ([]*github.Repository, *github.Response, error) {
	affiliations := make([]string, 0)

	if *c.Owned {
		affiliations = append(affiliations, "owner")
	}
	if *c.Collaborator {
		affiliations = append(affiliations, "collaborator")
	}
	if *c.OrgMember {
		affiliations = append(affiliations, "organization_member")
	}

	if len(affiliations) == 0 {
		return make([]*github.Repository, 0), &github.Response{}, nil
	}

	return c.client.Repositories.List(context.Background(), "", &github.RepositoryListOptions{
		ListOptions: github.ListOptions{
			Page:    page,
			PerPage: 100,
		},
		Affiliation: strings.Join(affiliations, ","),
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
	for i := range repos {
		repos[i] = starred[i].Repository
	}
	return repos, response, err
}
