package npm

import (
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/exp/slog"
)

type NPMLock struct {
	Name     string             `json:"name"`
	Version  string             `json:"version"`
	Packages map[string]Package `json:"packages"`
}

type Package struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Resolved     string            `json:"resolved"`
	Integrity    string            `json:"integrity"`
	Dependencies map[string]string `json:"dependencies"`
}

var ErrMissingLockFilePath = errors.New("missing lockfile path")

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

	// Parse the lock file.
	log.Info("Parsing lock file")
	lockFile, err := parseLockFile(args.FileName)
	if err != nil {
		return fmt.Errorf("failed to parse lock file: %w", err)
	}

	// Create output directory if required.
	if err = createOutputDirectory(); err != nil {
		return err
	}

	// Download targz files into directory.
	// Create the packages channel.
	packages := make(chan Package)
	var downloadsCompleted int64
	var fromCache int64

	// Drain the channel concurrently.
	concurrency := 4
	var wg sync.WaitGroup
	errors := make([]error, concurrency)
	for i := 0; i < concurrency; i++ {
		go func(i int) {
			wg.Add(1)
			defer wg.Done()
			for pkg := range packages {
				targetFileName, err := getFileName(pkg.Resolved)
				if err != nil {
					errors[i] = err
					break
				}
				if isAlreadyDownloaded(targetFileName, pkg.Integrity) {
					atomic.AddInt64(&fromCache, 1)
					continue
				}
				err = download(pkg.Resolved, targetFileName, pkg.Integrity)
				if err != nil {
					errors[i] = err
				}
				atomic.AddInt64(&downloadsCompleted, 1)
			}
		}(i)
	}

	// Filter the packages.
	var resolvedPackages []Package
	for _, pkg := range lockFile.Packages {
		// If there's no URL, skip.
		if pkg.Resolved == "" {
			continue
		}
		resolvedPackages = append(resolvedPackages, pkg)
	}
	downloadsTotal := len(resolvedPackages)

	// Push data into the channel, and notify when complete.
	done := make(chan struct{}, 1)
	go func() {
		for _, pkg := range resolvedPackages {
			packages <- pkg
		}
		close(packages)
		done <- struct{}{}
	}()
	log.Info("Downloading packages", slog.Int("total", downloadsTotal))

	// Display status.
	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()
exit:
	for {
		select {
		case <-done:
			break exit
		case <-ticker.C:
			log.Info("In progress", slog.Int("total", downloadsTotal), slog.Int("downloads", int(downloadsCompleted)), slog.Int("fromCache", int(fromCache)), slog.String("duration", time.Now().Sub(start).String()))
		}
	}
	wg.Wait()

	log.Info("Complete", slog.Int("total", downloadsTotal), slog.Int("downloads", int(downloadsCompleted)), slog.Int("fromCache", int(fromCache)), slog.String("duration", time.Now().Sub(start).String()))
	return nil
}

func getFileName(from string) (fileName string, err error) {
	u, err := url.Parse(from)
	if err != nil {
		return fileName, err
	}
	_, name := path.Split(u.Path)
	if name == "" {
		return fileName, fmt.Errorf("no filename present in path %q for URL %q", u.Path, from)
	}
	return path.Join("packages", name), nil
}

func hashFile(name string) (hash string, err error) {
	r, err := os.Open(name)
	if err != nil {
		return
	}
	defer r.Close()
	return hashReader(r)
}

func hashReader(r io.Reader) (hash string, err error) {
	hasher := sha512.New()
	if _, err = io.Copy(hasher, r); err != nil {
		return
	}
	return "sha512-" + base64.StdEncoding.EncodeToString(hasher.Sum(nil)), err
}

func validateFileHash(fileName string, expectedHash string) (err error) {
	actualHash, err := hashFile(fileName)
	if err != nil {
		return err
	}
	if expectedHash != actualHash {
		return fmt.Errorf("expected hash %q, got %q", expectedHash, actualHash)
	}
	return nil
}

func validateReaderHash(r io.Reader, expectedHash string) (err error) {
	actualHash, err := hashReader(r)
	if err != nil {
		return err
	}
	if expectedHash != actualHash {
		return fmt.Errorf("expected hash %q, got %q", expectedHash, actualHash)
	}
	return nil
}

func isAlreadyDownloaded(targetFileName, expectedHash string) bool {
	_, err := os.Stat(targetFileName)
	if !errors.Is(err, os.ErrNotExist) {
		err = validateFileHash(targetFileName, expectedHash)
		if err == nil {
			return true
		}
	}
	return false
}

func download(from, targetFileName, expectedHash string) (err error) {
	// Download the file.
	resp, err := http.Get(from)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("expected status OK for %q, but got %d", from, resp.StatusCode)
	}

	// Create the target.
	w, err := os.Create(targetFileName)
	if err != nil {
		return err
	}

	// Hash the file while we download it.
	hashR, hashW := io.Pipe()
	var writeErr error
	go func() {
		_, writeErr = io.Copy(io.MultiWriter(w, hashW), resp.Body)
		hashW.Close()
	}()
	err = validateReaderHash(hashR, expectedHash)
	if err != nil {
		return err
	}
	if writeErr != nil {
		return writeErr
	}

	return err
}

func createOutputDirectory() error {
	if _, err := os.Stat("packages"); os.IsNotExist(err) {
		return os.Mkdir("packages", 0770)
	}
	return nil
}

func parseLockFile(fileName string) (lockFile NPMLock, err error) {
	f, err := os.Open(fileName)
	if err != nil {
		return
	}
	err = json.NewDecoder(f).Decode(&lockFile)
	return
}
