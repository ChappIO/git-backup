package git_backup

import (
	"log"
	"net/url"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

type RepositorySource interface {
	GetName() string
	Test() error
	ListRepositories() ([]*Repository, error)
}

type Repository struct {
	GitURL   url.URL
	FullName string
}

func isBare(repo *git.Repository) bool {
	config, err := repo.Config()
	if err != nil {
		return false
	}

	return config.Core.IsBare
}

func (r *Repository) CloneInto(path string, bare bool) error {
	var auth http.AuthMethod
	if r.GitURL.User != nil {
		password, _ := r.GitURL.User.Password()
		auth = &http.BasicAuth{
			Username: r.GitURL.User.Username(),
			Password: password,
		}
	}
	gitRepo, err := git.PlainClone(path, bare, &git.CloneOptions{
		URL:      r.GitURL.String(),
		Auth:     auth,
		Progress: os.Stdout,
	})

	if err == git.ErrRepositoryAlreadyExists {
		// as the repo already exists, we just plain open it
		gitRepo, err = git.PlainOpen(path)
		if err != nil {
			return err
		}
	}

	var remotes []*git.Remote // we only create this variable here so we don't have to use := with the Remotes() call, which would create a new err
	remotes, err = gitRepo.Remotes()
	if err != nil {
		return err
	}

	// we iterate over all the remotes, and we fetch all their heads
	for _, remote := range remotes {
		err = remote.Fetch(&git.FetchOptions{
			Auth:     auth,
			Progress: os.Stdout,
			Tags:     git.AllTags,
			Force:    true,
			RefSpecs: []config.RefSpec{"+refs/heads/*:refs/heads/*"},
		})
		if err != nil && err != git.NoErrAlreadyUpToDate {
			break
		}
	}

	if !isBare(gitRepo) {
		var worktree *git.Worktree // don't wanna shadow err
		worktree, err = gitRepo.Worktree()
		if err != nil {
			return err
		}

		// this should probably be replaced with a Checkout call instead
		err = worktree.Pull(&git.PullOptions{
			Auth:     auth,
			Progress: os.Stdout,
		})
	}

	switch err {
	case transport.ErrEmptyRemoteRepository:
		log.Printf("%s is an empty repository", r.FullName)
		//  Empty repo does not need backup
		return nil
	case git.NoErrAlreadyUpToDate:
		log.Printf("No need to pull, %s is already up-to-date", r.FullName)
		return nil
	default:
		return err
	}
}
