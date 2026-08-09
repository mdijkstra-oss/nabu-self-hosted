package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func fixture(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func allKeys() map[string]string {
	return map[string]string{
		"OPENAI_API_KEY":     "sk-test",
		"ANTHROPIC_API_KEY":  "sk-test",
		"GEMINI_API_KEY":     "sk-test",
		"DEEPSEEK_API_KEY":   "sk-test",
		"OPENROUTER_API_KEY": "sk-test",
	}
}

const extendsModels = `models:
  strong:
    model: anthropic/claude-opus-5
    reasoning_effort: medium
  strong-planning:
    extends: strong
    prompt: nabu/modes/planning.md
  strong-execution:
    extends: strong
    prompt: nabu/modes/execution.md
`

const emptyAuthService = `local:
  endpoint: http://localhost:11434/v1
  protocol: openai-completions
`

func TestRun(t *testing.T) {
	cases := []struct {
		name      string
		models    string
		dragoman  string
		env       map[string]string
		exit      int
		stderrHas []string
		stderrNot []string
		lines     int
	}{
		{
			name:     "skeleton: openai preset with only OPENAI_API_KEY passes silently",
			models:   fixture(t, "models.openai.yaml"),
			dragoman: fixture(t, "dragoman.yaml"),
			env:      map[string]string{"OPENAI_API_KEY": "sk-test"},
			exit:     0,
			lines:    0,
		},
		{
			name:      "unparseable models yaml names the file, not a missing key",
			models:    "models: [unclosed\n\t",
			dragoman:  fixture(t, "dragoman.yaml"),
			env:       allKeys(),
			exit:      1,
			stderrHas: []string{"cannot parse", "models"},
			stderrNot: []string{"required by"},
		},
		{
			name:      "unknown prefix names the prefix and the model value",
			models:    "models:\n  strong:\n    model: acme/some-model\n",
			dragoman:  fixture(t, "dragoman.yaml"),
			env:       allKeys(),
			exit:      1,
			stderrHas: []string{`"acme"`, `"acme/some-model"`},
		},
		{
			name:      "model value without a slash cannot be routed",
			models:    "models:\n  strong:\n    model: claude-opus-5\n",
			dragoman:  fixture(t, "dragoman.yaml"),
			env:       allKeys(),
			exit:      1,
			stderrHas: []string{"cannot be routed", `"claude-opus-5"`},
		},
		{
			name:      "multi preset missing GEMINI_API_KEY names the variable and its prefix",
			models:    fixture(t, "models.multi.yaml"),
			dragoman:  fixture(t, "dragoman.yaml"),
			env:       map[string]string{"ANTHROPIC_API_KEY": "sk-test", "OPENAI_API_KEY": "sk-test"},
			exit:      1,
			stderrHas: []string{"GEMINI_API_KEY", `"gemini"`},
			stderrNot: []string{"ANTHROPIC_API_KEY", "OPENAI_API_KEY"},
		},
		{
			name:      "OPENAI_API_KEY is required by the embeddings service even off the openai preset",
			models:    fixture(t, "models.anthropic.yaml"),
			dragoman:  fixture(t, "dragoman.yaml"),
			env:       map[string]string{"ANTHROPIC_API_KEY": "sk-test"},
			exit:      1,
			stderrHas: []string{"OPENAI_API_KEY", "embeddings"},
			stderrNot: []string{"ANTHROPIC_API_KEY"},
		},
		{
			name:      "several missing variables are all reported in one run, one line each",
			models:    fixture(t, "models.multi.yaml"),
			dragoman:  fixture(t, "dragoman.yaml"),
			env:       map[string]string{},
			exit:      1,
			stderrHas: []string{"OPENAI_API_KEY", "ANTHROPIC_API_KEY", "GEMINI_API_KEY"},
			lines:     3,
		},
		{
			name:     "extends aliases contribute nothing beyond the extended entry's model",
			models:   extendsModels,
			dragoman: fixture(t, "dragoman.yaml"),
			env:      map[string]string{"ANTHROPIC_API_KEY": "sk-test", "OPENAI_API_KEY": "sk-test"},
			exit:     0,
			lines:    0,
		},
		{
			name:      "a variable set to the empty string counts as missing",
			models:    fixture(t, "models.openai.yaml"),
			dragoman:  fixture(t, "dragoman.yaml"),
			env:       map[string]string{"OPENAI_API_KEY": ""},
			exit:      1,
			stderrHas: []string{"OPENAI_API_KEY"},
		},
		{
			name:     "values are never verified: garbage keys pass silently",
			models:   fixture(t, "models.multi.yaml"),
			dragoman: fixture(t, "dragoman.yaml"),
			env: map[string]string{
				"OPENAI_API_KEY":    "definitely-not-a-key",
				"ANTHROPIC_API_KEY": "12345",
				"GEMINI_API_KEY":    " ",
			},
			exit:  0,
			lines: 0,
		},
		{
			name:      "unparseable dragoman yaml names that file",
			models:    fixture(t, "models.openai.yaml"),
			dragoman:  "openai: [unclosed\n\t",
			env:       allKeys(),
			exit:      1,
			stderrHas: []string{"cannot parse", "dragoman"},
		},
		{
			name:     "matched http service with empty auth requires no variable",
			models:   "models:\n  strong:\n    model: local/llama\n",
			dragoman: emptyAuthService,
			env:      map[string]string{"OPENAI_API_KEY": "sk-test"},
			exit:     0,
			lines:    0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			modelsPath := filepath.Join(dir, "models.yaml")
			dragomanPath := filepath.Join(dir, "dragoman.yaml")
			if err := os.WriteFile(modelsPath, []byte(tc.models), 0o644); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(dragomanPath, []byte(tc.dragoman), 0o644); err != nil {
				t.Fatal(err)
			}

			var stderr strings.Builder
			exit := run([]string{"-models", modelsPath, "-dragoman", dragomanPath}, tc.env, &stderr)

			if exit != tc.exit {
				t.Errorf("exit = %d, want %d; stderr:\n%s", exit, tc.exit, stderr.String())
			}
			out := stderr.String()
			if tc.exit == 0 && out != "" {
				t.Errorf("success must print nothing, got:\n%s", out)
			}
			for _, want := range tc.stderrHas {
				if !strings.Contains(out, want) {
					t.Errorf("stderr misses %q:\n%s", want, out)
				}
			}
			for _, banned := range tc.stderrNot {
				if strings.Contains(out, banned) {
					t.Errorf("stderr must not contain %q:\n%s", banned, out)
				}
			}
			if tc.lines > 0 || tc.exit == 0 {
				got := len(strings.Split(strings.TrimRight(out, "\n"), "\n"))
				if out == "" {
					got = 0
				}
				if got != tc.lines {
					t.Errorf("stderr has %d lines, want %d:\n%s", got, tc.lines, out)
				}
			}
		})
	}
}

func TestMissingModelsFileNamesIt(t *testing.T) {
	dir := t.TempDir()
	dragomanPath := filepath.Join(dir, "dragoman.yaml")
	if err := os.WriteFile(dragomanPath, []byte(fixture(t, "dragoman.yaml")), 0o644); err != nil {
		t.Fatal(err)
	}
	var stderr strings.Builder
	missing := filepath.Join(dir, "absent.yaml")
	exit := run([]string{"-models", missing, "-dragoman", dragomanPath}, allKeys(), &stderr)
	if exit != 1 {
		t.Errorf("exit = %d, want 1", exit)
	}
	if !strings.Contains(stderr.String(), missing) {
		t.Errorf("stderr misses the file path:\n%s", stderr.String())
	}
}
