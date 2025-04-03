package git_backup

import (
	"log"
	"net/url"
	"slices"
	"strings"

	"github.com/xanzy/go-gitlab"
)

type GitLabConfig struct {
	URL         string   `yaml:"url,omitempty"`
	JobName     string   `yaml:"job_name"`
	AccessToken string   `yaml:"access_token"`
	Starred     *bool    `yaml:"starred,omitempty"`
	Member      *bool    `yaml:"member,omitempty"`
	Owned       *bool    `yaml:"owned,omitempty"`
	Exclude     []string `yaml:"exclude,omitempty"`
	client      *gitlab.Client
}

func (g *GitLabConfig) GetName() string {
	return g.JobName
}

func (g *GitLabConfig) Test() error {
	user, _, err := g.client.Users.CurrentUser()
	if err != nil {
		return err
	}
	log.Printf("Authenticated as with gitlab as: %s", user.Username)
	return nil
}

func (g *GitLabConfig) ListRepositories() ([]*Repository, error) {
	out := make(map[string]*Repository, 0)

	if *g.Starred {
		if repos, err := g.getAllRepos(&gitlab.ListProjectsOptions{Starred: boolPointer(true)}); err != nil {
			return nil, err
		} else {
			for _, repo := range repos {
				out[repo.FullName] = repo
			}
		}
	}

	if *g.Owned {
		if repos, err := g.getAllRepos(&gitlab.ListProjectsOptions{Owned: boolPointer(true)}); err != nil {
			return nil, err
		} else {
			for _, repo := range repos {
				out[repo.FullName] = repo
			}
		}
	}

	if *g.Member {
		if repos, err := g.getAllRepos(&gitlab.ListProjectsOptions{Membership: boolPointer(true)}); err != nil {
			return nil, err
		} else {
			for _, repo := range repos {
				out[repo.FullName] = repo
			}
		}
	}

	outSlice := make([]*Repository, 0, len(out))
	for _, repository := range out {
		isExcluded := slices.ContainsFunc(g.Exclude, func(s string) bool {
			if strings.EqualFold(s, repository.FullName) {
				return true
			}

			if strings.Contains(s, "/") {
				return false
			}

			repoFullName := repository.FullName

			repoOwner := repoFullName[:strings.Index(repoFullName, "/")]
			return strings.EqualFold(s, repoOwner)
		})
		if isExcluded {
			log.Printf("Skipping excluded repository: %s", repository.FullName)
		} else {
			outSlice = append(outSlice, repository)
		}
	}

	return outSlice, nil
}

func (g *GitLabConfig) getAllRepos(opts *gitlab.ListProjectsOptions) ([]*Repository, error) {
	out := make([]*Repository, 0)
	for i := 1; true; i++ {
		opts.ListOptions.Page = i
		repos, _, err := g.getRepos(opts)
		if err != nil {
			return out, err
		}
		for _, repo := range repos {
			gitUrl, err := url.Parse(repo.HTTPURLToRepo)
			if err != nil {
				return out, err
			}
			gitUrl.User = url.UserPassword("git", g.AccessToken)
			out = append(out, &Repository{
				GitURL:   *gitUrl,
				FullName: repo.PathWithNamespace,
			})
		}
		if len(repos) == 0 {
			break
		}
	}
	return out, nil
}

func (g *GitLabConfig) getRepos(opts *gitlab.ListProjectsOptions) ([]*gitlab.Project, *gitlab.Response, error) {
	opts.ListOptions.PerPage = 100
	opts.Simple = boolPointer(true)
	return g.client.Projects.ListProjects(opts)
}

func (g *GitLabConfig) setDefaults() {
	if g.Member == nil {
		g.Member = boolPointer(true)
	}
	if g.Owned == nil {
		g.Owned = boolPointer(true)
	}
	if g.Starred == nil {
		g.Starred = boolPointer(true)
	}
	if g.JobName == "" {
		g.JobName = "GitLab"
	}
	if g.URL == "" {
		g.client, _ = gitlab.NewClient(g.AccessToken)
	} else {
		client, err := gitlab.NewClient(g.AccessToken, gitlab.WithBaseURL(g.URL))
		if err != nil {
			panic(err)
		}
		g.client = client
	}
}
