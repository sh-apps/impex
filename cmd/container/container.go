package container

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"
	"unicode"

	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"

	"log/slog"
)

type Arguments struct {
	FileName string
	Log      *slog.Logger
}

func Run(args Arguments) error {
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

	// Create output directory if required.
	if err = createOutputDirectory(); err != nil {
		return err
	}

	// Download OCI container images.
	downloads := strings.Split(string(data), "\n")
	var errs error
	var downloadsComplete int
	for _, container := range downloads {
		if len(container) == 0 {
			downloadsComplete++
			continue
		}
		if strings.HasPrefix(container, "#") {
			log.Info("Skipping", slog.String("name", container), slog.Int("total", len(downloads)))
			downloadsComplete++
			continue
		}
		log.Info("Downloading", slog.String("name", container), slog.Int("total", len(downloads)))
		err := download(container)
		if err != nil {
			errs = errors.Join(err)
		}
		downloadsComplete++
	}

	log.Info("Complete", slog.Int("total", len(downloads)), slog.String("duration", time.Now().Sub(start).String()))
	return errs
}

func download(name string) (err error) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return fmt.Errorf("failed to create CLI client: %w", err)
	}

	// Pull the image.
	reader, err := cli.ImagePull(ctx, name, types.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("%s: failed to pull image: %w", name, err)
	}
	_, err = io.Copy(os.Stdout, reader)
	if err != nil {
		return fmt.Errorf("%s: failed to copy image: %w", name, err)
	}

	// Save the output.
	r, err := cli.ImageSave(ctx, []string{name})
	defer r.Close()
	if err != nil {
		return fmt.Errorf("%s: failed to save image: %w", name, err)
	}

	// Create the target.
	targetFileName := path.Join("package/containers", getFileName(name))
	w, err := os.Create(targetFileName)
	if err != nil {
		return fmt.Errorf("%s: failed to create target file %q: %w", name, targetFileName, err)
	}

	// Copy data.
	_, err = io.Copy(w, r)
	if err != nil {
		return fmt.Errorf("%s: failed to copy data: %w", name, err)
	}

	return nil
}

func getFileName(name string) string {
	var sb strings.Builder
	for _, c := range name {
		if !unicode.IsLetter(c) && !unicode.IsDigit(c) {
			sb.WriteRune('_')
			continue
		}
		sb.WriteRune(c)
	}
	return sb.String()
}

func createOutputDirectory() error {
	if _, err := os.Stat("package/containers"); os.IsNotExist(err) {
		return os.MkdirAll("package/containers", 0770)
	}
	return nil
}
