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

Set `OPENAI_API_KEY` in `.env`, then start the stack:

```sh
docker compose up
```

The first run builds every service from source and takes several minutes. Once the stack is up, open <http://localhost:8090>. The app offers to create a first project.

Two checks run before any service starts. `chancery validate` reads the seeded configuration and reports every fault in it — a `MODELS` naming no such table, an agent naming an alias the table lacks, a prompt including a file that is not there. A second check stops the stack when `OPENAI_API_KEY` is missing. Either one failing stops the stack rather than leaving a service to crash and restart.

The other provider keys are checked when they are used. A request for a model whose provider key is unset fails with an error naming that variable. To see every provider and whether its key is set:

```sh
docker compose exec proxy wget -qO- http://dragoman:8080/services
```

## Choosing models

> [!IMPORTANT]
> The embeddings service supports OpenAI only. No setting directs it to another provider, so `OPENAI_API_KEY` is required even when every model tier runs on Anthropic, Gemini, or DeepSeek.

`MODELS` in `.env` names the models yaml chancery reads. nabu-prompts ships five, and `models.openai.yaml` is the default:

```sh
MODELS=models.anthropic.yaml
```

The others are `models.gemini.yaml`, `models.deepseek.yaml` and `models.multi.yaml`. Each needs its provider's API key, and `multi` needs several. A model whose key is missing fails when it is first used, naming the variable.

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

Project files are stored in the Docker volume `projects`. They survive `docker compose down` and are removed only by `docker compose down -v`.

`STORAGE_DATA` accepts a `./`-relative or absolute host path and stores project files in that directory instead. The directory must be writable by uid 65532, the fixed unprivileged user the storage service runs as.

## Updating

```sh
git pull
docker compose build
docker compose up
```

Each build fetches the `main` branch of every service repository. Builds are therefore not reproducible: nothing pins a service to a release, and the same commands can produce different images on two machines. Set the `*_REPO` variables to local checkouts to build from a fixed tree.

## Configuration reference

Every setting is read from `.env`. `.env.example` documents each one next to its default.

| Variable | Default | What it does |
| --- | --- | --- |
| `OPENAI_API_KEY` | empty | Required. Used by the embeddings service, and by `models.openai.yaml` and `models.multi.yaml` |
| `ANTHROPIC_API_KEY` | empty | Required by `models.anthropic.yaml` and `models.multi.yaml` |
| `GEMINI_API_KEY` | empty | Required by `models.gemini.yaml` and `models.multi.yaml` |
| `DEEPSEEK_API_KEY` | empty | Required by `models.deepseek.yaml` |
| `OPENROUTER_API_KEY` | unset | Reachable only from a models yaml of your own naming `openrouter/` models |
| `MODELS` | `models.openai.yaml` | Which models yaml chancery reads. A bare name is one of the five nabu-prompts ships; an absolute path is a table you mounted |
| `NABU_PORT` | `8090` | The stack's only published host port |
| `STORAGE_DATA` | `projects` | Named volume, or a host path for project files |
| `NABU_FRONTEND_REPO`, `NABU_STORAGE_REPO`, `NABU_EMBEDDINGS_REPO`, `NABU_PROMPTS_REPO`, `CHANCERY_REPO`, `DRAGOMAN_REPO` | unset | Builds a service from a local working copy instead of its GitHub repository |

## See also

- [nabu-frontend](https://github.com/mdijkstra-oss/nabu-frontend) — the web app
- [nabu-storage](https://github.com/mdijkstra-oss/nabu-storage) — project files and the sync API
- [nabu-embeddings](https://github.com/mdijkstra-oss/nabu-embeddings) — the `/embeddings` route
- [nabu-prompts](https://github.com/mdijkstra-oss/nabu-prompts) — prompt files and the models tables
- [chancery](https://github.com/mdijkstra-oss/chancery) — the agent gateway behind `/llm`
- [dragoman](https://github.com/mdijkstra-oss/dragoman) — provider routing beneath chancery
