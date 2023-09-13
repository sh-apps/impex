package git

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"context"

	"github.com/go-git/go-git/v5"

	"log/slog"
)

type Arguments struct {
	FileName string
	Log      *slog.Logger
}

func Run(args Arguments) error {
	ctx := context.Background()
	start := time.Now()

	// Create log.
	log := args.Log
	if log == nil {
		log = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	}

	// Parse the input file.
	log.Info("Parsing input file")
	data, err := os.ReadFile(args.FileName)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}

	// Download git repos.
	downloads := strings.Split(string(data), "\n")
	var errs error
	var downloadsComplete int
	for _, repo := range downloads {
		if len(repo) == 0 {
			downloadsComplete++
			continue
		}
		if strings.HasPrefix(repo, "#") {
			log.Info("Skipping", slog.String("name", repo), slog.Int("total", len(downloads)))
			downloadsComplete++
			continue
		}
		log.Info("Downloading", slog.String("name", repo), slog.Int("total", len(downloads)))
		err := download(ctx, repo)
		if err != nil {
			errs = errors.Join(err)
		}
		downloadsComplete++
	}

	log.Info("Complete", slog.Int("total", len(downloads)), slog.String("duration", time.Now().Sub(start).String()))
	return errs
}

func download(ctx context.Context, gitURL string) (err error) {
	// Get the name.
	u, err := url.Parse(gitURL)
	if err != nil {
		return fmt.Errorf("failed to parse url: %w", err)
	}
	host := u.Hostname()
	if u.Hostname() == "" {
		host = "localhost"
	}

	// Create the target.
	targetPath := path.Join("package/git", host, strings.ToLower(u.Path))
	if err = createOutputDirectory(targetPath); err != nil {
		return err
	}

	// Clone the repo.
	_, err = git.PlainClone("package/git/", false, &git.CloneOptions{
		URL:      gitURL,
		Progress: os.Stdout,
	})

	return err
}

func createOutputDirectory(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return os.MkdirAll(dir, 0770)
	}
	return nil
}
