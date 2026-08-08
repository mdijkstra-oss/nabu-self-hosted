# Nabu self-hosted

Nabu self-hosted runs the whole Nabu stack — the web app, project storage, embeddings, and the LLM gateway with its prompt config — as one Docker Compose stack behind a single published port.

You clone this repository, put your API keys in `.env`, run `docker compose up`, and open one URL. Every service is built from its own repository at build time; this repository holds only the stack definition.

## Prerequisites

- Docker with Compose v2.33 or newer — the stack builds its preflight check from another service's image, which older Compose versions cannot resolve. Check with `docker compose version`.
- An OpenAI API key. It is always required, whichever models you pick, because the embeddings service uses it. Other providers need their own key only when your model selection names them.
- Network access to github.com (the builds fetch each service's source from there) and to the model providers.

## Install

```sh
git clone https://github.com/mdijkstra-oss/nabu-self-hosted.git
cd nabu-self-hosted
cp .env.example .env
```

Open `.env` and set `OPENAI_API_KEY`. Then:

```sh
docker compose up
```

The first run builds every service from source, so it takes a few minutes. When the stack is up, open <http://localhost:8090> — the app lands you in a welcome project.

Before anything serves, a preflight check verifies that every API key your model selection needs is present. When one is missing, the stack stops instead of starting half-broken, and the preflight's output names the missing variable.

## Choosing models

`MODELS` in `.env` selects a preset: `openai` (the default), `anthropic`, `gemini`, `deepseek`, or `multi`. Each preset needs its provider's API key set — `multi` needs several; the preflight tells you exactly which.

`MODELS_FILE`, when set to a host path, uses your own models yaml instead of any preset and `MODELS` is ignored. A path that does not point at a readable file stops the stack rather than silently falling back.

A changed model selection takes effect on the next `docker compose up`; no rebuild is needed.

## Where data lives

Your projects live in the Docker volume `projects`. They survive `docker compose down`; only `docker compose down -v` deletes them.

Set `STORAGE_DATA` in `.env` to a `./`-relative or absolute host path to keep project files in a directory instead. The directory must be writable by uid 65532, the fixed unprivileged user the storage service runs as.

## Updating

```sh
git pull
docker compose build
docker compose up
```

Always run the full `docker compose build`: the preflight check captures the LLM gateway's provider table from the image built alongside it, so rebuilding one service alone can leave the two out of step until the next full build.

## Configuration reference

All knobs live in `.env`; `.env.example` documents each next to its default.

| Variable | Default | What it does |
| --- | --- | --- |
| `OPENAI_API_KEY` | empty | Always required; used for embeddings and by the `openai` and `multi` presets |
| `ANTHROPIC_API_KEY` | empty | Needed by the `anthropic` and `multi` presets |
| `GEMINI_API_KEY` | empty | Needed by the `gemini` and `multi` presets |
| `DEEPSEEK_API_KEY` | empty | Needed by the `deepseek` preset |
| `OPENROUTER_API_KEY` | unset | Only reachable from a `MODELS_FILE` yaml that names `openrouter/` models |
| `MODELS` | `openai` | Selects a models preset |
| `MODELS_FILE` | unset | Host path to your own models yaml; wins over `MODELS` |
| `NABU_PORT` | `8090` | Host port the stack publishes — its only one |
| `STORAGE_DATA` | `projects` | Named volume, or a host path for project files |
| `NABU_FRONTEND_REPO`, `NABU_STORAGE_REPO`, `NABU_EMBEDDINGS_REPO`, `NABU_PROMPTS_REPO`, `CHANCERY_REPO`, `DRAGOMAN_REPO` | unset | Build a service from a local working copy instead of its GitHub repository, for side-by-side development |
