# Nabu self-hosted

Nabu self-hosted runs the Nabu stack as one Docker Compose project behind a single published port: the web app, project storage, embeddings, and the LLM gateway with its prompt configuration.

Every service is built from its own repository at build time. This repository holds the stack definition and the preflight check that guards it.

## Prerequisites

- Docker with Compose v2.33 or newer. The stack builds its preflight check from another service's image, which earlier versions cannot resolve. Verify with `docker compose version`.
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

A preflight check runs before any service accepts traffic. It verifies that every API key the model selection requires is present, and stops the stack when one is missing, naming the variable.

## Choosing models

> [!IMPORTANT]
> The embeddings service supports OpenAI only. No setting directs it to another provider, so `OPENAI_API_KEY` is required even when every model tier runs on Anthropic, Gemini, or DeepSeek.

`MODELS` in `.env` selects a preset: `openai` (the default), `anthropic`, `gemini`, `deepseek`, or `multi`. Each preset requires its provider's API key. The `multi` preset requires several, and the preflight names them.

`MODELS_FILE` takes a host path to a models yaml and overrides `MODELS`. A path that does not resolve to a readable file stops the stack.

A changed model selection applies on the next `docker compose up` and requires no rebuild.

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

Run the full `docker compose build`. The preflight check captures the LLM gateway's provider table from the image built alongside it, so rebuilding a single service can leave the two out of step until the next full build.

## Configuration reference

Every setting is read from `.env`. `.env.example` documents each one next to its default.

| Variable | Default | What it does |
| --- | --- | --- |
| `OPENAI_API_KEY` | empty | Required. Used by the embeddings service and by the `openai` and `multi` presets |
| `ANTHROPIC_API_KEY` | empty | Required by the `anthropic` and `multi` presets |
| `GEMINI_API_KEY` | empty | Required by the `gemini` and `multi` presets |
| `DEEPSEEK_API_KEY` | empty | Required by the `deepseek` preset |
| `OPENROUTER_API_KEY` | unset | Reachable only from a `MODELS_FILE` yaml naming `openrouter/` models |
| `MODELS` | `openai` | Selects a models preset |
| `MODELS_FILE` | unset | Host path to a models yaml. Overrides `MODELS` |
| `NABU_PORT` | `8090` | The stack's only published host port |
| `STORAGE_DATA` | `projects` | Named volume, or a host path for project files |
| `NABU_FRONTEND_REPO`, `NABU_STORAGE_REPO`, `NABU_EMBEDDINGS_REPO`, `NABU_PROMPTS_REPO`, `CHANCERY_REPO`, `DRAGOMAN_REPO` | unset | Builds a service from a local working copy instead of its GitHub repository |

## See also

- [nabu-frontend](https://github.com/mdijkstra-oss/nabu-frontend) — the web app
- [nabu-storage](https://github.com/mdijkstra-oss/nabu-storage) — project files and the sync API
- [nabu-embeddings](https://github.com/mdijkstra-oss/nabu-embeddings) — the `/embeddings` route
- [nabu-prompts](https://github.com/mdijkstra-oss/nabu-prompts) — prompt files and the models presets
- [chancery](https://github.com/mdijkstra-oss/chancery) — the agent gateway behind `/llm`
- [dragoman](https://github.com/mdijkstra-oss/dragoman) — provider routing beneath chancery
