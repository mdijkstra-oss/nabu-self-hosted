package main

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
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

func main() {
	storageURL := os.Getenv("STORAGE_URL")
	if storageURL == "" {
		storageURL = "http://storage:8080"
	}
	os.Exit(run(storageURL, os.Stderr))
}

func run(storageURL string, stderr io.Writer) int {
	empty, err := storageIsEmpty(storageURL)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if !empty {
		return 0
	}
	if err := seedWelcomeProject(storageURL); err != nil {
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

func seedWelcomeProject(storageURL string) error {
	projectID, err := newUUID()
	if err != nil {
		return err
	}

	body, err := json.Marshal(command{Action: "WriteFile", Path: "welcome.md", Content: welcomeText})
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
