package git_backup

import (
	"log"
	"net/url"
	"os"

	"github.com/go-git/go-git/v5"
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

func isBare(repo *git.Repository) (bool, error) {
	config, err := repo.Config()
	if err != nil {
		return false, err
	}

	return config.Core.IsBare, nil
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
		// Pull instead of clone
		if gitRepo, err = git.PlainOpen(path); err == nil {
			// we need to check whether it's a bare repo or not.
			// if not we should pull, if it is then pull won't work
			if isBare, bErr := isBare(gitRepo); bErr == nil && !isBare {
				if w, wErr := gitRepo.Worktree(); wErr != nil {
					err = wErr
				} else {
					err = w.Pull(&git.PullOptions{
						Auth:     auth,
						Progress: os.Stdout,
					})
				}
			}
		}
	}

	switch err {
	case transport.ErrEmptyRemoteRepository:
		log.Printf("%s is an empty repository", r.FullName)
		//  Empty repo does not need backup
		return nil
	default:
		return err
	case git.NoErrAlreadyUpToDate:
		log.Printf("No need to pull, %s is already up-to-date", r.FullName)
		// Already up to date on current branch, still need to refresh other branches
		fallthrough
	case nil:
		// No errors, continue
		err = gitRepo.Fetch(&git.FetchOptions{
			Auth:     auth,
			Progress: os.Stdout,
			Tags:     git.AllTags,
			Force:    true,
		})
	}

	switch err {
	case git.NoErrAlreadyUpToDate:
		log.Printf("No need to fetch, %s is already up-to-date", r.FullName)
		return nil
	default:
		return err
	}
}
