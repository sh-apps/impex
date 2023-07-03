package vsix

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"golang.org/x/exp/slog"
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

	// Download vsix files.
	downloads := strings.Split(string(data), "\n")
	var errs error
	var downloadsComplete int
	for _, pkg := range downloads {
		if len(pkg) == 0 {
			continue
		}
		log.Info("Downloading", slog.String("name", pkg), slog.Int("total", len(downloads)))
		err := download(pkg)
		if err != nil {
			errs = errors.Join(err)
		}
		downloadsComplete++
	}

	log.Info("Complete", slog.Int("total", len(downloads)), slog.String("duration", time.Now().Sub(start).String()))
	return errs
}

func download(name string) (err error) {
	// Download the file.
	parts := strings.SplitN(name, ".", 2)
	publisher, name := parts[0], parts[1]
	from := fmt.Sprintf("https://%s.gallery.vsassets.io/_apis/public/gallery/publisher/%s/extension/%s/latest/assetbyname/Microsoft.VisualStudio.Services.VSIXPackage",
		url.PathEscape(publisher),
		url.PathEscape(publisher),
		url.PathEscape(name))

	resp, err := http.Get(from)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("expected status OK for %q, but got %d", from, resp.StatusCode)
	}

	// Create the target.
	targetFileName := path.Join("extensions", name+".vsix")
	w, err := os.Create(targetFileName)
	if err != nil {
		return err
	}

	// Copy data.
	_, err = io.Copy(w, resp.Body)
	return err
}

func createOutputDirectory() error {
	if _, err := os.Stat("extensions"); os.IsNotExist(err) {
		return os.Mkdir("extensions", 0770)
	}
	return nil
}
