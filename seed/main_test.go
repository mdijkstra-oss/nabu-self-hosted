package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

var commandPath = regexp.MustCompile(
	`^/commands/[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`,
)

type recordedPost struct {
	Method string
	Path   string
	Body   []byte
}

type fakeStorage struct {
	*httptest.Server
	projects []string
	posts    []recordedPost
}

func newFakeStorage(t *testing.T, projects ...string) *fakeStorage {
	t.Helper()
	fake := &fakeStorage{projects: projects}
	fake.Server = httptest.NewServer(http.HandlerFunc(fake.handle))
	t.Cleanup(fake.Close)
	return fake
}

func (f *fakeStorage) handle(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodGet && r.URL.Path == "/queries/projects":
		items := make([]map[string]string, 0, len(f.projects))
		for _, id := range f.projects {
			items = append(items, map[string]string{"id": id, "updatedAt": "2026-08-08T00:00:00Z"})
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"items": items, "total": len(items), "page": 1, "page_size": 50,
		})
	case r.Method == http.MethodPost && strings.HasPrefix(r.URL.Path, "/commands/"):
		body, _ := io.ReadAll(r.Body)
		f.posts = append(f.posts, recordedPost{Method: r.Method, Path: r.URL.Path, Body: body})
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	default:
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
	}
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func TestSeedsWhenStorageHoldsZeroProjects(t *testing.T) {
	fake := newFakeStorage(t)

	if code := run(fake.URL, io.Discard); code != 0 {
		t.Fatalf("run() = %d, want 0", code)
	}
	if len(fake.posts) != 1 {
		t.Fatalf("storage received %d writes, want exactly 1", len(fake.posts))
	}

	post := fake.posts[0]
	if post.Method != http.MethodPost {
		t.Errorf("method = %q, want POST", post.Method)
	}
	if !commandPath.MatchString(post.Path) {
		t.Errorf("path = %q, want /commands/{uuidv4}", post.Path)
	}

	var cmd struct {
		Action  string `json:"action"`
		Path    string `json:"path"`
		Content string `json:"content"`
	}
	if err := json.Unmarshal(post.Body, &cmd); err != nil {
		t.Fatalf("body is not JSON: %v: %s", err, post.Body)
	}
	if cmd.Action != "WriteFile" {
		t.Errorf("action = %q, want WriteFile", cmd.Action)
	}
	if cmd.Path != "welcome.md" {
		t.Errorf("path = %q, want welcome.md", cmd.Path)
	}
	if !strings.HasPrefix(cmd.Content, "# Welcome to Nabu!") {
		t.Errorf("content begins %.40q, want prefix %q", cmd.Content, "# Welcome to Nabu!")
	}
}

func TestDoesNothingWhenAnyProjectExists(t *testing.T) {
	fake := newFakeStorage(t, "0b0e9646-88fc-4a48-9c1a-6b04a2f0f3a4")

	if code := run(fake.URL, io.Discard); code != 0 {
		t.Fatalf("run() = %d, want 0", code)
	}
	if len(fake.posts) != 0 {
		t.Fatalf("storage received %d writes, want none", len(fake.posts))
	}
}

// A project whose welcome.md was deleted still lists as a project, and the
// project list is the whole empty signal — so nothing is recreated.
func TestNeverRecreatesAfterWelcomeFileDeleted(t *testing.T) {
	fake := newFakeStorage(t, "4e6f7462-6f6f-4b0d-8a1c-000000000001")

	if code := run(fake.URL, io.Discard); code != 0 {
		t.Fatalf("run() = %d, want 0", code)
	}
	if len(fake.posts) != 0 {
		t.Fatalf("storage received %d writes, want none", len(fake.posts))
	}
}

func TestQueryErrorExitsNonZeroWithResponseOnStderr(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
	}))
	t.Cleanup(server.Close)

	var stderr strings.Builder
	if code := run(server.URL, &stderr); code == 0 {
		t.Fatal("run() = 0, want non-zero")
	}
	if !strings.Contains(stderr.String(), "internal error") {
		t.Errorf("stderr = %q, want the storage response in it", stderr.String())
	}
}

func TestWriteErrorExitsNonZeroWithResponseOnStderr(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			writeJSON(w, http.StatusOK, map[string]any{
				"items": []any{}, "total": 0, "page": 1, "page_size": 50,
			})
			return
		}
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid projectId"})
	}))
	t.Cleanup(server.Close)

	var stderr strings.Builder
	if code := run(server.URL, &stderr); code == 0 {
		t.Fatal("run() = 0, want non-zero")
	}
	if !strings.Contains(stderr.String(), "invalid projectId") {
		t.Errorf("stderr = %q, want the storage response in it", stderr.String())
	}
}

func TestUnreachableStorageExitsNonZero(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	url := server.URL
	server.Close()

	var stderr strings.Builder
	if code := run(url, &stderr); code == 0 {
		t.Fatal("run() = 0, want non-zero")
	}
	if stderr.Len() == 0 {
		t.Error("stderr is empty, want the connection error")
	}
}

func TestMalformedProjectsResponseExitsNonZero(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "not json")
	}))
	t.Cleanup(server.Close)

	var stderr strings.Builder
	if code := run(server.URL, &stderr); code == 0 {
		t.Fatal("run() = 0, want non-zero")
	}
}
