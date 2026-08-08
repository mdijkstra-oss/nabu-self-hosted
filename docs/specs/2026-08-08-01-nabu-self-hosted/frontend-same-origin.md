# Frontend same-origin URLs

nabu-frontend's three backend address vars learn one rule — a value starting with `/` is same-origin, resolved against the page's own origin at runtime — so one built bundle reaches its backends through the [proxy](proxy.md)'s single origin whether the browser opened `localhost` or a LAN IP.

## Contract

One value grammar covers `VITE_API_HOST`, `VITE_LLM_HOST`, and `VITE_EMBEDDINGS_URL`; each value takes one of three forms.

| Value form | Meaning |
| --- | --- |
| no leading `/` (bare `host[:port]` for the API var, full URL for the LLM and embeddings vars) | today's meaning, byte-for-byte unchanged for the API and LLM vars; for the embeddings var the meaning arrives with its new name |
| leading `/` | same-origin: a path prefix joined with the caller's path (API and LLM vars), or a root-relative URL used verbatim (embeddings var); no trailing `/` |
| unset or `""` | the localhost default baked into each env module — never same-origin |

Empty can never mean same-origin: `getEnv` (`app/lib/utils/env.ts`) coalesces `""` to the fallback because a Docker build ARG declared but not passed arrives as the empty string, so same-origin always requires a real `/`-prefixed value.

There are two patch sites. The first is `app/lib/server/env.ts`: today `getApiUrl` and `getWsUrl` always prepend a scheme to the raw `VITE_API_HOST` value (default `localhost:8080`), which makes a path prefix unusable; with the patch, a `/`-prefixed value makes them build `<scheme>://<window.location.host><value><path>`, the scheme mirroring `window.location.protocol` exactly as today — `https:` gives `https`/`wss`, anything else gives `http`/`ws`.

| `VITE_API_HOST` | Page origin | `getApiUrl("/queries/projects")` | `getWsUrl("/ws/p1")` |
| --- | --- | --- | --- |
| `localhost:8080` (or unset) | `http://localhost:5173` | `http://localhost:8080/queries/projects` | `ws://localhost:8080/ws/p1` |
| `/api` | `http://192.168.1.20:8090` | `http://192.168.1.20:8090/api/queries/projects` | `ws://192.168.1.20:8090/api/ws/p1` |
| `/api` | `https://nabu.example` | `https://nabu.example/api/queries/projects` | `wss://nabu.example/api/ws/p1` |

The builders emit absolute URLs rather than relative paths because `new WebSocket(url)` (`app/lib/server/sync/websocket.ts`) needs an absolute `ws(s)://` URL — browser support for relative WebSocket URLs is recent and uneven — and deriving from `window.location` keeps both HTTP and WS callers on the same rule.

The second patch site is the embeddings address: `VITE_EMBEDDINGS_HOST`, a base the client suffixes with `/embeddings` at its single call site, becomes `VITE_EMBEDDINGS_URL`, the full endpoint URL, default `http://localhost:8082/embeddings` — the same effective target as today.

The service exposes exactly one endpoint and every caller threads the value untouched to that one call site (`app/lib/embeddings/client.ts`), so the base-plus-append indirection had no second consumer; with the append gone, the var's value is the request URL — subject only to the trailing-slash strip below — and a `/`-prefixed value is simply a root-relative URL the browser resolves against the page origin.

The getter in `app/lib/embeddings/env.ts`, the `baseUrl` parameter threading it to the client, the Dockerfile's build ARG, and the repo's `.env.example` rename with the var in one change; a deployment still setting `VITE_EMBEDDINGS_HOST` is silently ignored — `getEnv` falls back to the default — an accepted pre-release break, with no fallback read of the old name.

| Var | Where the value becomes a URL | Baked value in the stack | Browser request |
| --- | --- | --- | --- |
| `VITE_LLM_HOST` | `app/lib/agent/client/fetch.ts` `buildUrl` (`app/lib/agent/env.ts`'s `getLlmUrl` has no callers) | `/llm` | `POST /llm/<agent-path>` |
| `VITE_EMBEDDINGS_URL` | used directly by `app/lib/embeddings/client.ts` | `/embeddings` | `POST /embeddings` |

Which public paths the proxy answers is [proxy.md](proxy.md)'s contract, and the actual baked values are [compose.md](compose.md)'s.

Side effects: none — `getApiUrl` and `getWsUrl` are pure string builders; the network calls happen in the consumers (`app/lib/server/sync/websocket.ts`, `app/lib/server/sync/commands.ts`, `app/lib/server/api/queries.ts`).

Enforcement: the grammar lives only where each var is already consumed — `server/env.ts` is the single module all three server-side callers import their URLs from, and the LLM and embeddings getters are the only sources for theirs — nothing else in the app re-derives a backend URL, so no caller can bypass the rule.

Every caller path starts with `/` (`/ws/{projectId}`, `/commands/{projectId}`, `/queries/projects`), and the consumers strip one trailing `/` from a `/`-prefixed value — `/api/` means `/api`, and `/embeddings/` means `/embeddings`, which would otherwise miss the proxy's exact route and land in the SPA fallback — while bare-host and full-URL values pass through byte-for-byte, keeping the first grammar row's promise.

## Prior art

In-repo: the concat-based LLM getters (`getLlmUrl`, `buildUrl`) are already prefix-tolerant, and the patch extends that established pattern to `server/env.ts`, the one builder that forces a scheme today.

Online: path prefixes behind one reverse-proxy origin, with the SPA using origin-relative URLs, is the standard single-origin deployment pattern for static bundles.

Rejected: a runtime config endpoint (`/config.json` fetched at boot) — a second request and a loading gate to deliver three constants. Rejected: baking `window.location.host` at build time — impossible, a static bundle has no runtime host until a browser opens it.

## Tests

1. Skeleton — built with `VITE_API_HOST=/api` and served through the stack's proxy, the app loads at the published origin (localhost or LAN IP), fetches the project list from `/api/queries/projects`, and syncs the welcome project's files at connect over `wss://<origin>/api/ws/{projectId}`, with no other origin contacted.

2. Contract, riskiest first:
   Given `VITE_API_HOST=/api` and a page at `https://192.168.1.20:8443`, when `getWsUrl("/ws/p1")` and `getApiUrl("/queries/projects")` run, then they return `wss://192.168.1.20:8443/api/ws/p1` and `https://192.168.1.20:8443/api/queries/projects`.
   Given the same `/api` value on an `http:` page, when the builders run, then the schemes are `ws` and `http` — mirroring `window.location.protocol`.
   Given `VITE_API_HOST=localhost:8080` (bare) on an `http:` page, when the builders run, then the output is exactly today's `http://localhost:8080/...` and `ws://localhost:8080/...` — the regression pin for existing deployments; no current test pins this behavior (verified: no test file references `getApiUrl`/`getWsUrl`), so this pin is new.
   Given `VITE_API_HOST=""`, when the builders run, then behavior equals unset (the `localhost:8080` default) — pinning the `getEnv` coalescing the "empty is never same-origin" clause rests on.
   Given `VITE_EMBEDDINGS_URL=/embeddings`, when the embeddings client fetches, then the request goes to `/embeddings` on the page origin.
   Given `VITE_EMBEDDINGS_URL` unset, when the client fetches, then the request goes to `http://localhost:8082/embeddings` — the same target the old base-plus-append produced, the regression pin across the rename.
   Given `VITE_EMBEDDINGS_URL=/embeddings/` with a trailing slash, when the client fetches, then the request path is `/embeddings` — stripped, so it cannot fall through to the SPA fallback.
   Given value `/api` and caller path `/ws/p1`, when joined, then the result contains `/api/ws/p1` with a single slash at the join; and given the value `/api/` with its trailing slash, then the result is identical.

3. Isolation — vitest (`npm test` runs `vitest run`; `vite.config.ts` configures no jsdom or browser test environment), cases-as-table style per `app/lib/embeddings/hash.test.ts`, `window.location` stubbed via `vi.stubGlobal` and the vars via `vi.stubEnv`; pure string assertions, no network, no socket opened.
