package git

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/pkg/errors"
)

type Git struct {
	repository *git.Repository
}

func Get() (*Git, error) {
	// We instantiate a new repository targeting the given path (the .git folder)
	repository, err := git.PlainOpen(".")
	if err != nil {
		return nil, errors.Wrap(err, "(git) cannot instantiate a new repository")
	}

	return &Git{
		repository: repository,
	}, nil
}

func (g *Git) Head() (*plumbing.Reference, error) {
	ref, err := g.repository.Head()
	if err != nil {
		return nil, errors.Wrap(err, "failed to run git log")
	}

	return ref, nil
}

func (g *Git) Tags() (storer.ReferenceIter, error) {
	tags, err := g.repository.Tags()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get git tags")
	}

	return tags, nil
}

func (g *Git) IsWorktreeClean() (bool, error) {
	worktree, err := g.repository.Worktree()
	if err != nil {
		return false, errors.Wrap(err, "failed to get worktree")
	}

	status, err := worktree.Status()
	if err != nil {
		return false, errors.Wrap(err, "failed to run git status --porcelain")
	}

	return status.IsClean(), nil
}
