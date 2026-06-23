package git_backup

import (
	"context"
	"fmt"
	"log"
	"maps"
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
	Repositories []string `yaml:"repositories,omitempty"`
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
		if isExcluded(c.Exclude, *repo.FullName) {
			log.Printf("Skipping excluded repository: %s", *repo.FullName)
			continue
		}

		gitUrl, err := url.Parse(*repo.CloneURL)
		if err != nil {
			return out, err
		}
		gitUrl.User = url.UserPassword("github", c.AccessToken)

		out = append(out, &Repository{
			FullName: *repo.FullName,
			GitURL:   *gitUrl,
		})

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
	all := make(map[string]*github.Repository, 0)

	// fetch the configured repos
	if len(c.Repositories) > 0 {
		if repos, err := c.getRepoList(c.Repositories); err != nil {
			return nil, err
		} else {
			for _, repo := range repos {
				all[*repo.FullName] = repo
			}
		}
	}

	// use discovery mechanisms
	var err error

	for repos, response, apiErr := c.discoverRepos(1); true; repos, response, apiErr = c.discoverRepos(response.NextPage) {
		if apiErr != nil {
			err = apiErr
			break
		} else {
			for _, repo := range repos {
				all[*repo.FullName] = repo
			}
		}

		if len(repos) == 0 || response.NextPage == 0 {
			break
		}
	}
	if err != nil {
		return nil, err
	}

	if *c.Starred {
		for repos, response, apiErr := c.getStarredRepos(1); true; repos, response, apiErr = c.getStarredRepos(response.NextPage) {
			if apiErr != nil {
				err = apiErr
				break
			} else {
				for _, repo := range repos {
					all[*repo.FullName] = repo
				}
			}

			if len(repos) == 0 || response.NextPage == 0 {
				break
			}
		}
	}

	return slices.Collect(maps.Values(all)), err
}

func (c *GithubConfig) getRepoList(repos []string) ([]*github.Repository, error) {

	ghRepos := make([]*github.Repository, 0, len(repos))

	for _, repo := range repos {
		parts := strings.Split(repo, "/")

		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid repository name '%v' must be of schema owner/repo", repo)
		}

		ghRepo, _, err := c.client.Repositories.Get(context.Background(), parts[0], parts[1])
		if err != nil {
			return nil, err
		}
		ghRepos = append(ghRepos, ghRepo)
	}

	return ghRepos, nil
}

func (c *GithubConfig) discoverRepos(page int) ([]*github.Repository, *github.Response, error) {
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
