package git_backup

import (
	"log"
	"maps"
	"net/url"
	"slices"

	gitlab "gitlab.com/gitlab-org/api/client-go"
)

type GitLabConfig struct {
	URL          string   `yaml:"url,omitempty"`
	JobName      string   `yaml:"job_name"`
	AccessToken  string   `yaml:"access_token"`
	Starred      *bool    `yaml:"starred,omitempty"`
	Member       *bool    `yaml:"member,omitempty"`
	Owned        *bool    `yaml:"owned,omitempty"`
	Exclude      []string `yaml:"exclude,omitempty"`
	Repositories []string `yaml:"repositories,omitempty"`
	client       *gitlab.Client
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

func (g *GitLabConfig) ListRepositories() ([]*Repository, error) {
	repos, err := g.getAllRepos()
	if err != nil {
		return nil, err
	}

	out := make([]*Repository, 0, len(repos))
	for _, repo := range repos {
		if isExcluded(g.Exclude, repo.PathWithNamespace) {
			log.Printf("Skipping excluded repository: %s", repo.PathWithNamespace)
			continue
		}

		gitUrl, err := url.Parse(repo.HTTPURLToRepo)
		if err != nil {
			return nil, err
		}
		gitUrl.User = url.UserPassword("git", g.AccessToken)

		out = append(out, &Repository{
			GitURL:   *gitUrl,
			FullName: repo.PathWithNamespace,
		})
	}

	return out, nil
}

func (g *GitLabConfig) getAllRepos() ([]*gitlab.Project, error) {
	all := make(map[string]*gitlab.Project, 0)

	// fetch the configured repos
	if len(g.Repositories) > 0 {
		if repos, err := g.getRepoList(g.Repositories); err != nil {
			return nil, err
		} else {
			for _, repo := range repos {
				all[repo.PathWithNamespace] = repo
			}
		}
	}

	// use discovery mechanisms
	if *g.Starred {
		if repos, err := g.discoverRepos(&gitlab.ListProjectsOptions{Starred: boolPointer(true)}); err != nil {
			return nil, err
		} else {
			for _, repo := range repos {
				all[repo.PathWithNamespace] = repo
			}
		}
	}

	if *g.Owned {
		if repos, err := g.discoverRepos(&gitlab.ListProjectsOptions{Owned: boolPointer(true)}); err != nil {
			return nil, err
		} else {
			for _, repo := range repos {
				all[repo.PathWithNamespace] = repo
			}
		}
	}

	if *g.Member {
		if repos, err := g.discoverRepos(&gitlab.ListProjectsOptions{Membership: boolPointer(true)}); err != nil {
			return nil, err
		} else {
			for _, repo := range repos {
				all[repo.PathWithNamespace] = repo
			}
		}
	}

	return slices.Collect(maps.Values(all)), nil
}

func (g *GitLabConfig) getRepoList(repos []string) ([]*gitlab.Project, error) {
	glProjects := make([]*gitlab.Project, 0, len(repos))

	for _, repo := range repos {
		project, _, err := g.client.Projects.GetProject(repo, &gitlab.GetProjectOptions{License: boolPointer(false), Statistics: boolPointer(false)})
		if err != nil {
			return nil, err
		}
		glProjects = append(glProjects, project)

	}
	return glProjects, nil

}

func (g *GitLabConfig) discoverRepos(opts *gitlab.ListProjectsOptions) ([]*gitlab.Project, error) {

	opts.ListOptions.PerPage = 100
	opts.Simple = boolPointer(true)

	repos := make([]*gitlab.Project, 0, 100)

	// paginate
	for i := 1; true; i++ {
		opts.ListOptions.Page = i
		projects, _, err := g.client.Projects.ListProjects(opts)
		if err != nil {
			return nil, err
		}

		if len(projects) == 0 {
			break
		}
		repos = append(repos, projects...)

	}
	return repos, nil
}
