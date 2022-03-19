package git_backup

import (
	"github.com/xanzy/go-gitlab"
	"log"
	"net/url"
	"strings"
)

type GitLabConfig struct {
	URL         string `yaml:"url,omitempty"`
	JobName     string `yaml:"job_name"`
	AccessToken string `yaml:"access_token"`
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

	if repos, err := g.getAllRepos(&gitlab.ListProjectsOptions{Starred: boolPointer(true)}); err != nil {
		return nil, err
	} else {
		for _, repo := range repos {
			out[repo.FullName] = repo
		}
	}

	if repos, err := g.getAllRepos(&gitlab.ListProjectsOptions{Owned: boolPointer(true)}); err != nil {
		return nil, err
	} else {
		for _, repo := range repos {
			out[repo.FullName] = repo
		}
	}

	if repos, err := g.getAllRepos(&gitlab.ListProjectsOptions{Membership: boolPointer(true)}); err != nil {
		return nil, err
	} else {
		for _, repo := range repos {
			out[repo.FullName] = repo
		}
	}

	outSlice := make([]*Repository, 0, len(out))
	for _, repository := range out {
		outSlice = append(outSlice, repository)
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
			nameParts := strings.Split(repo.NameWithNamespace, "/")
			for i2, part := range nameParts {
				nameParts[i2] = strings.TrimSpace(part)
			}
			out = append(out, &Repository{
				GitURL:   *gitUrl,
				FullName: strings.Join(nameParts, "/"),
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