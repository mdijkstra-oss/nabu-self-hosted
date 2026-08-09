package main

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// The empty signal is the project list, not any particular file: a user who
// deletes welcome.md but keeps a project is never re-seeded.
const welcomeText = `# Welcome to Nabu!

Nabu is your personal writing space. Everything here is a plain Markdown
file, stored by your own self-hosted server — no accounts, no cloud, just
your words on your machine.

This project was created for you on first start, so you would land somewhere
to write instead of on an empty screen. There is nothing special about it:
it is an ordinary project like any you create yourself.

Edit this file freely, delete it, or start a fresh document — it is all
yours. Happy writing!
`

type projectPage struct {
	Items []json.RawMessage `json:"items"`
	Total int               `json:"total"`
}

type command struct {
	Action  string `json:"action"`
	Path    string `json:"path"`
	Content string `json:"content"`
}

type seedFile struct {
	path    string
	content string
}

func main() {
	storageURL := os.Getenv("STORAGE_URL")
	if storageURL == "" {
		storageURL = "http://storage:8080"
	}
	os.Exit(run(storageURL, os.Stderr))
}

func run(storageURL string, stderr io.Writer) int {
	// SEED_DIR is read here rather than in main so run keeps the signature
	// the pinned seed tests call it with.
	files, err := loadSeedFiles(os.Getenv("SEED_DIR"))
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	empty, err := storageIsEmpty(storageURL)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if !empty {
		return 0
	}
	if err := seedProject(storageURL, files); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}

func storageIsEmpty(storageURL string) (bool, error) {
	resp, err := http.Get(storageURL + "/queries/projects")
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("GET /queries/projects: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return false, fmt.Errorf("GET /queries/projects: %s: %s", resp.Status, body)
	}

	var page projectPage
	if err := json.Unmarshal(body, &page); err != nil {
		return false, fmt.Errorf("GET /queries/projects: %w: %s", err, body)
	}
	return len(page.Items) == 0 && page.Total == 0, nil
}

// Unset SEED_DIR keeps the embedded welcome document, so every deployment
// without the e2e mount seeds exactly what it did before SEED_DIR existed.
func loadSeedFiles(dir string) ([]seedFile, error) {
	if dir == "" {
		return []seedFile{{path: "welcome.md", content: welcomeText}}, nil
	}

	// os.ReadDir sorts by filename, which makes the seed order deterministic.
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("SEED_DIR: %w", err)
	}

	files := make([]seedFile, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			return nil, fmt.Errorf("SEED_DIR: %s: %q is a directory, the corpus must be flat", dir, entry.Name())
		}
		if !entry.Type().IsRegular() {
			continue
		}
		if !validFilename(entry.Name()) {
			return nil, fmt.Errorf("SEED_DIR: %s: %q is outside storage's filename grammar", dir, entry.Name())
		}
		content, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("SEED_DIR: %w", err)
		}
		files = append(files, seedFile{path: entry.Name(), content: string(content)})
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("SEED_DIR: %s contains no regular files", dir)
	}
	return files, nil
}

// Mirrors storage's grammar (nabu-storage internal/lib/utils/id.go,
// ValidFilePath) so a bad name fails the boot instead of a mid-seed write.
func validFilename(name string) bool {
	if name == "" || strings.HasPrefix(name, ".") || strings.Contains(name, "..") {
		return false
	}
	for _, r := range name {
		if !isSafeFilenameChar(r) {
			return false
		}
	}
	return true
}

func isSafeFilenameChar(r rune) bool {
	if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' {
		return true
	}
	switch r {
	case '-', '_', '.', ' ', '(', ')', '\'', ',':
		return true
	}
	return false
}

func seedProject(storageURL string, files []seedFile) error {
	projectID, err := newUUID()
	if err != nil {
		return err
	}
	for _, file := range files {
		if err := writeFile(storageURL, projectID, file); err != nil {
			return err
		}
	}
	return nil
}

func writeFile(storageURL, projectID string, file seedFile) error {
	body, err := json.Marshal(command{Action: "WriteFile", Path: file.path, Content: file.content})
	if err != nil {
		return err
	}

	resp, err := http.Post(storageURL+"/commands/"+projectID, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	answer, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("POST /commands/%s: %w", projectID, err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("POST /commands/%s: %s: %s", projectID, resp.Status, answer)
	}
	return nil
}

// Random UUIDv4 from crypto/rand, version and variant bits per RFC 4122 —
// storage canonicalizes the id with uuid.Parse, so this is all it requires.
func newUUID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}
