package git

import (
	"fmt"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

func Copy(accessToken, src, tgt string) error {
	dir := os.TempDir()
	repo, err := git.PlainClone(dir, false, &git.CloneOptions{
		URL: src,
		Auth: &http.BasicAuth{
			Username: "git",
			Password: accessToken,
		},
		Progress: os.Stdout,
	})
	if err != nil {
		return fmt.Errorf("failed to clone: %w", err)
	}
	return repo.Push(&git.PushOptions{
		RemoteURL: tgt,
		Auth: &http.BasicAuth{
			Username: "git",
			Password: accessToken,
		},
		Force:      true,
		FollowTags: true,
		Progress:   os.Stdout,
	})
}
