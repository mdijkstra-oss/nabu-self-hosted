# Nabu self-hosted

Nabu self-hosted runs the Nabu stack as one Docker Compose project behind a single published port: the web app, project storage, embeddings, and the LLM gateway with its prompt configuration.

Every service is built from its own repository at build time. This repository holds the stack definition.

## Prerequisites

- Docker with Compose v2.
- An OpenAI API key. The embeddings service requires it regardless of which models are selected. Other providers require their own key only when the model selection names them.
- Network access to github.com, which the builds fetch each service's source from, and to the model providers.

## Install

```sh
git clone https://github.com/mdijkstra-oss/nabu-self-hosted.git
cd nabu-self-hosted
cp .env.example .env
```

Set `OPENAI_API_KEY` and `PROJECT_DIR` in `.env`. `PROJECT_DIR` is the directory your project files are written to, and it has to exist:

```sh
mkdir -p ~/nabu-notes
```

To start with a project already in it, copy the sample instead:

```sh
cp -R sample-project ~/nabu-notes
```

Then start the stack:

```sh
docker compose up
```

The first run builds every service from source and takes several minutes. Once the stack is up, open <http://localhost:8090>. An empty `PROJECT_DIR` means the app offers to create a first project.

Three checks run before any service starts. `chancery validate` reads the seeded configuration and reports every fault in it — a `MODELS` naming no such table, an agent naming an alias the table lacks, a prompt including a file that is not there. Two more stop the stack when `OPENAI_API_KEY` is missing and when `PROJECT_DIR` is not writable. Any one failing stops the stack rather than leaving a service to crash and restart.

The other provider keys are checked when they are used. A request for a model whose provider key is unset fails with an error naming that variable. To see every provider and whether its key is set:

```sh
docker compose exec proxy wget -qO- http://dragoman:8080/services
```

## Choosing models

> [!IMPORTANT]
> The embeddings service supports OpenAI only. No setting directs it to another provider, so `OPENAI_API_KEY` is required even when every model tier runs on Anthropic, Gemini, or DeepSeek.

`MODELS` in `.env` names the models yaml chancery reads. The five it can name live in [nabu-prompts/config](https://github.com/mdijkstra-oss/nabu-prompts/tree/main/config), and [`models.openai.yaml`](https://github.com/mdijkstra-oss/nabu-prompts/blob/main/config/models.openai.yaml) is the default:

```sh
MODELS=models.anthropic.yaml
```

The example above is [`models.anthropic.yaml`](https://github.com/mdijkstra-oss/nabu-prompts/blob/main/config/models.anthropic.yaml). The remaining three are [`models.gemini.yaml`](https://github.com/mdijkstra-oss/nabu-prompts/blob/main/config/models.gemini.yaml), [`models.deepseek.yaml`](https://github.com/mdijkstra-oss/nabu-prompts/blob/main/config/models.deepseek.yaml) and [`models.multi.yaml`](https://github.com/mdijkstra-oss/nabu-prompts/blob/main/config/models.multi.yaml). Each needs its provider's API key, and `multi` needs several. A model whose key is missing fails when it is first used, naming the variable.

To run your own table instead, mount it into chancery and give `MODELS` its absolute path:

```yaml
services:
  chancery:
    volumes:
      - ./my-models.yaml:/etc/nabu/models.yaml:ro
```

```sh
MODELS=/etc/nabu/models.yaml
```

Switching needs no rebuild and no reseed. Every table nabu-prompts ships sits in the config volume at once, so changing `MODELS` is a restart:

```sh
MODELS=models.anthropic.yaml docker compose up -d chancery
```

## Where data lives

`PROJECT_DIR` is a directory on your disk, and nothing else. There is no default and no name with a special meaning: a value without a `/` is rejected, and so is a path that does not exist. A `./`-relative path resolves against this repository's directory rather than your shell's working directory.

The directory must be writable by uid 65532, the fixed unprivileged user storage runs as. On Docker Desktop that is automatic. On Linux it takes `sudo chown -R 65532:65532 <dir>`.

Inside it, one directory per project, named by the project's UUID. The listing is read fresh on every request, so a project directory copied in by hand appears like any other — which is all `sample-project/` is. Nothing seeds it for you and nothing rewrites it, so the files are yours to move, back up, or edit outside the app.

Point `PROJECT_DIR` at a throwaway directory to get a stack that starts clean:

```sh
PROJECT_DIR=$(mktemp -d) docker compose up
```

Don't point it into this repository. Storage writes there, so your notes would land under version control and be lost on the next `git pull`.

## Updating

```sh
git pull
docker compose build
docker compose up
```

Each build fetches the `main` branch of every service repository. Builds are therefore not reproducible: nothing pins a service to a release, and the same commands can produce different images on two machines. Set the `*_REPO` variables to local checkouts to build from a fixed tree.

## Development

`docker compose up` builds each service into an image, so a one-line change costs a rebuild. The `Procfile` next to this file runs the same stack as five native processes instead, each rebuilt and restarted the moment its own repository changes.

It expects every repository checked out beside this one, with chancery and dragoman one level up:

```
.
├── chancery
├── dragoman
└── nabu
    ├── nabu-embeddings
    ├── nabu-frontend
    ├── nabu-prompts
    ├── nabu-self-hosted
    └── nabu-storage
```

Beyond that it needs Go, Node, [overmind](https://github.com/DarthSim/overmind), [watchexec](https://github.com/watchexec/watchexec) and Caddy:

```sh
brew install overmind watchexec caddy
make dev
```

`make dev` stops before starting anything if a tool is missing, a repository is not checked out, or one of the ports below is already in use, and names every fault it found at once. `make check` reports the same thing without starting anything, and covers the container stack too. `make` on its own lists every target.

The app is on <http://localhost:5173>. `.env` supplies the API keys and `PROJECT_DIR`, exactly as it does for compose, so a stack you have already configured needs nothing added.

There is no proxy. Each service publishes its own port and the app addresses them directly:

| Process | Port | Rebuilds when |
| --- | --- | --- |
| `frontend` | 5173 | never — Vite hot-reloads the module |
| `storage` | 8080 | `nabu-storage/**/*.go` |
| `chancery` | 8081 | `chancery/**/*.go`, and any prompt or models table in `nabu-prompts/config` |
| `embeddings` | 8082 | `nabu-embeddings/Caddyfile` |
| `dragoman` | 8083 | `dragoman/{cmd,internal}/**/*.{go,yaml}` |

dragoman moves to 8083 here. In the container stack it shares 8080 with storage, which is only possible because each has its own network namespace.

Nothing negotiates these numbers. Each backend is told to allow `http://localhost:5173` and no other origin, so an app served from anywhere else is refused by every one of them. That is why a port already in use is a hard stop rather than a warning, and why the app refuses to start on the next free port the way a Vite dev server otherwise would.

Two differences from the compose stack are worth knowing, because they are the parts this arrangement cannot exercise. Requests are cross-origin rather than same-origin, so the app exercises the CORS path that production never reaches. And the browser talks to storage, chancery and embeddings directly, so the Caddy routing in `Caddyfile` is not in the picture at all. [nabu-e2e](https://github.com/mdijkstra-oss/nabu-e2e) runs against the compose stack and covers both.

It runs in the foreground, and Ctrl-C stops all five. Closing the window instead leaves `.overmind.sock` behind, which the next `make dev` removes after checking that nothing is behind it.

`overmind status` lists the five processes and their pids. `overmind connect chancery` attaches to one process's terminal, which is the readable way to follow a single service or run a debugger against it. `overmind restart storage` restarts one by hand.

## Configuration reference

Every setting is read from `.env`. `.env.example` documents each one next to its default.

| Variable | Default | What it does |
| --- | --- | --- |
| `OPENAI_API_KEY` | empty | Required. Used by the embeddings service, and by `models.openai.yaml` and `models.multi.yaml` |
| `ANTHROPIC_API_KEY` | empty | Required by `models.anthropic.yaml` and `models.multi.yaml` |
| `GEMINI_API_KEY` | empty | Required by `models.gemini.yaml` and `models.multi.yaml` |
| `DEEPSEEK_API_KEY` | empty | Required by `models.deepseek.yaml` |
| `OPENROUTER_API_KEY` | unset | Reachable only from a models yaml of your own naming `openrouter/` models |
| `MODELS` | `models.openai.yaml` | Which models yaml chancery reads. A bare name is one of the five in [nabu-prompts/config](https://github.com/mdijkstra-oss/nabu-prompts/tree/main/config); an absolute path is a table you mounted |
| `NABU_PORT` | `8090` | The stack's only published host port |
| `PROJECT_DIR` | none — required | Host directory project files are written to. Must exist and be writable by uid 65532 |
| `NABU_FRONTEND_REPO`, `NABU_STORAGE_REPO`, `NABU_EMBEDDINGS_REPO`, `NABU_PROMPTS_REPO`, `CHANCERY_REPO`, `DRAGOMAN_REPO` | unset | Builds a service from a local working copy instead of its GitHub repository |

## See also

- [nabu-frontend](https://github.com/mdijkstra-oss/nabu-frontend) — the web app
- [nabu-storage](https://github.com/mdijkstra-oss/nabu-storage) — project files and the sync API
- [nabu-embeddings](https://github.com/mdijkstra-oss/nabu-embeddings) — the `/embeddings` route
- [nabu-prompts](https://github.com/mdijkstra-oss/nabu-prompts) — prompt files and the models tables
- [chancery](https://github.com/mdijkstra-oss/chancery) — the agent gateway behind `/llm`
- [dragoman](https://github.com/mdijkstra-oss/dragoman) — provider routing beneath chancery
