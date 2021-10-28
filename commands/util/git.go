package util

import (
	"fmt"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/tenderly/tenderly-cli/git"
)

func GetCommitish() *string {
	g, err := git.Get()
	if err != nil {
		return nil
	}

	ref, err := g.Head()
	if err != nil {
		return nil
	}

	hash := ref.Hash()
	if hash.IsZero() {
		dirty := ".dirty"
		return &dirty
	}

	clean, err := g.IsWorktreeClean()
	if err != nil {
		return nil
	}

	commitHash := hash.String()
	if !clean {
		dirtyCommit := fmt.Sprintf("%s.dirty", commitHash)
		return &dirtyCommit
	}

	name := ref.Name()
	if name.IsTag() {
		tag := name.Short()
		return &tag
	}

	tags, err := g.Tags()
	if err != nil {
		// Ignore failure
		return &commitHash
	}

	var tag *string
	_ = tags.ForEach(func(reference *plumbing.Reference) error {
		if reference.Hash().String() == commitHash {
			val := reference.Name().Short()
			tag = &val
		}
		return nil
	})
	if tag != nil {
		return tag
	}

	return &commitHash
}
