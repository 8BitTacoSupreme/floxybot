package canon

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Updater handles downloading canon snapshots from the backend.
type Updater struct {
	backendURL string
	canonDir   string
	client     *http.Client
}

func NewUpdater(backendURL, canonDir string) *Updater {
	return &Updater{
		backendURL: backendURL,
		canonDir:   canonDir,
		client:     &http.Client{Timeout: 120 * time.Second},
	}
}

// SnapshotPath returns the local path for the canon snapshot file.
func (u *Updater) SnapshotPath() string {
	return filepath.Join(u.canonDir, "canon.gob")
}

// ETagPath returns the path to the cached ETag file.
func (u *Updater) ETagPath() string {
	return filepath.Join(u.canonDir, ".etag")
}

// NeedsUpdate checks whether a fresh download is needed (max once per 24h).
func (u *Updater) NeedsUpdate() bool {
	info, err := os.Stat(u.SnapshotPath())
	if err != nil {
		return true // no local snapshot
	}
	return time.Since(info.ModTime()) > 24*time.Hour
}

// Download fetches the latest snapshot from the backend.
// Uses ETag for conditional download (304 Not Modified).
func (u *Updater) Download() error {
	if err := os.MkdirAll(u.canonDir, 0o755); err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodGet, u.backendURL+"/canon/latest", nil)
	if err != nil {
		return err
	}

	// Send cached ETag for conditional request.
	if etag, err := os.ReadFile(u.ETagPath()); err == nil {
		req.Header.Set("If-None-Match", string(etag))
	}

	resp, err := u.client.Do(req)
	if err != nil {
		return fmt.Errorf("downloading canon: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotModified {
		fmt.Println("Canon is up to date.")
		return nil
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("canon download HTTP %d", resp.StatusCode)
	}

	// Save snapshot.
	f, err := os.Create(u.SnapshotPath())
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		return err
	}

	// Save ETag.
	if etag := resp.Header.Get("ETag"); etag != "" {
		os.WriteFile(u.ETagPath(), []byte(etag), 0o644)
	}

	fmt.Println("Canon updated.")
	return nil
}
