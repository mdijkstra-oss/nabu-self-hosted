// Preflight proves every API key the stack will need exists before any other
// service starts. Contract: docs/specs/2026-08-08-01-nabu-self-hosted/preflight.md.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// nabu-embeddings injects this variable into every upstream request
// unconditionally, so it is required regardless of the models yaml.
const embeddingsVar = "OPENAI_API_KEY"

const embeddingsReason = "the embeddings service"

// modelEntry reads a strict subset of chancery's models.yaml field set
// (chancery/internal/prompts/types.go); every other field is ignored so this
// check can never be stricter about them than chancery itself.
type modelsFile struct {
	Models map[string]modelEntry `yaml:"models"`
}

type modelEntry struct {
	Model string `yaml:"model"`
}

// service reads a strict subset of dragoman's service entry
// (dragoman/internal/config/config.go); the document root is the service map.
type service struct {
	Endpoint string `yaml:"endpoint"`
	Auth     string `yaml:"auth"`
}

func main() {
	os.Exit(run(os.Args[1:], environ(os.Environ()), os.Stderr))
}

func environ(pairs []string) map[string]string {
	env := make(map[string]string, len(pairs))
	for _, pair := range pairs {
		name, value, _ := strings.Cut(pair, "=")
		env[name] = value
	}
	return env
}

func run(args []string, env map[string]string, stderr io.Writer) int {
	flags := flag.NewFlagSet("preflight", flag.ContinueOnError)
	flags.SetOutput(stderr)
	modelsPath := flags.String("models", "/config/models.yaml", "resolved models yaml the stack hands chancery")
	dragomanPath := flags.String("dragoman", "/etc/preflight/dragoman.yaml", "dragoman service-table yaml")
	if err := flags.Parse(args); err != nil {
		return 1
	}

	var models modelsFile
	if err := parse(*modelsPath, &models); err != nil {
		fmt.Fprintf(stderr, "preflight: %v\n", err)
		return 1
	}
	var services map[string]service
	if err := parse(*dragomanPath, &services); err != nil {
		fmt.Fprintf(stderr, "preflight: %v\n", err)
		return 1
	}

	failures := check(models.Models, services, env)
	for _, failure := range failures {
		fmt.Fprintf(stderr, "preflight: %s\n", failure)
	}
	if len(failures) > 0 {
		return 1
	}
	return 0
}

func parse(path string, target any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("cannot read %s: %v", path, err)
	}
	if err := yaml.Unmarshal(data, target); err != nil {
		return fmt.Errorf("cannot parse %s: %v", path, err)
	}
	return nil
}

// check returns one line per problem, every problem in one run: routing and
// exec failures per offending model value, then one line per missing variable.
func check(models map[string]modelEntry, services map[string]service, env map[string]string) []string {
	var failures []string
	demanded := []string{embeddingsVar}
	reasons := map[string][]string{embeddingsVar: {embeddingsReason}}

	for _, alias := range sortedKeys(models) {
		value := models[alias].Model
		if value == "" {
			continue
		}
		prefix, _, routable := strings.Cut(value, "/")
		matched, known := services[prefix]
		if !routable || !known {
			failures = append(failures, fmt.Sprintf("model %q cannot be routed: no dragoman service matches prefix %q", value, prefix))
			continue
		}
		if isSpawned(matched.Endpoint) {
			failures = append(failures, fmt.Sprintf("model %q needs service %q, which spawns a CLI (endpoint %q) and cannot serve in this stack", value, prefix, matched.Endpoint))
			continue
		}
		if matched.Auth == "" {
			continue
		}
		reason := fmt.Sprintf("model prefix %q", prefix)
		if _, seen := reasons[matched.Auth]; !seen {
			demanded = append(demanded, matched.Auth)
		}
		reasons[matched.Auth] = appendUnique(reasons[matched.Auth], reason)
	}

	for _, name := range demanded {
		if env[name] != "" {
			continue
		}
		failures = append(failures, fmt.Sprintf("%s is unset or empty, required by %s", name, strings.Join(reasons[name], " and ")))
	}
	return failures
}

// isSpawned mirrors dragoman's scheme cut (dragoman/internal/base/backend.go).
func isSpawned(endpoint string) bool {
	scheme, _, found := strings.Cut(endpoint, ":")
	return found && scheme == "exec"
}

func sortedKeys(models map[string]modelEntry) []string {
	keys := make([]string, 0, len(models))
	for key := range models {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func appendUnique(list []string, item string) []string {
	for _, existing := range list {
		if existing == item {
			return list
		}
	}
	return append(list, item)
}
