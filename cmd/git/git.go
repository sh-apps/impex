package git

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/google/go-github/v55/github"

	"log/slog"
)

type Arguments struct {
	FileName    string
	AccessToken string
	// Domain is the domain of the Github instance.
	// github.com - For public Github
	// github.example.com - For Github Enterprise
	Domain string
	Log    *slog.Logger
}

func Import(args Arguments) (err error) {
	client := github.NewClient(nil).WithAuthToken(args.AccessToken)
	if args.Domain != "" {
		client, err = client.WithEnterpriseURLs(args.Domain, args.Domain)
		if err != nil {
			return fmt.Errorf("failed to set domain: %w", err)
		}
	}
	_, _, err = client.Repositories.List(context.Background(), "", nil)
	return
}

func Export(args Arguments) error {
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
		err := download(repo, args.AccessToken)
		if err != nil {
			errs = errors.Join(err)
		}
		downloadsComplete++
	}

	log.Info("Complete", slog.Int("total", len(downloads)), slog.String("duration", time.Now().Sub(start).String()))
	return errs
}

func download(gitURL, accessToken string) (err error) {
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

	// Clone the repo.
	_, err = git.PlainClone(targetPath, false, &git.CloneOptions{
		URL: gitURL,
		Auth: &http.BasicAuth{
			Username: "git",
			Password: accessToken,
		},
		Progress: os.Stdout,
	})

	return err
}
