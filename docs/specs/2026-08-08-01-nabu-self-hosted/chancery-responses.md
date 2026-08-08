# chancery: `/responses` suffix on agent routes

OpenAI SDK clients append `/responses` to their `base_url`, so a client pointed at `http://chancery:8081/qual-coder` posts `/qual-coder/responses` and today gets `404 unknown agent: qual-coder/responses`. Chancery accepts the suffix, making every agent path a valid stock-SDK `base_url`. The nabu-self-hosted proxy path `/llm/<agent>/responses` rides on this — see [proxy.md](proxy.md).

## Contract

Accepted path forms on the wildcard `POST /*` route:

| path form | resolves to |
| --- | --- |
| `POST /<agent-path>` | that agent (unchanged) |
| `POST /<agent-path>.<model>` | that named model (unchanged) |
| `POST /<agent-path>/responses` | the agent, only when the literal path resolves nothing |
| `POST /<agent-path>.<model>/responses` | the named model, same rule — the suffix strips off the full path, so variants come along free |

Precedence: the exact path is tried first; only when that misses and the path ends in `/responses` is the suffix stripped and resolution tried once more. An agent literally named `responses` (config `research/responses.md` → route `/research/responses`) therefore keeps its own path — the exact match wins before any stripping.

A double miss is a `404` carrying whatever error the stripped reference gets from `ResolveAgent` today — `unknown agent: unknown` for a missing agent, the no-model-named message for `<agent>.<missing-model>` — so the strip changes only which reference the error names (`qual-coder`, not `qual-coder/responses`), because the agent path is what the SDK's `base_url` addressed and the suffix is SDK plumbing.

After a suffixed resolve, every downstream use of the request's agent identity carries the stripped reference: the `endpoint` field in the `request forwarded` and error logs (`chat.go` passes `urlPath` to both), and `Identity.Agent`, which travels to the backend as `X-Agent`. Logged as the resolved agent path, not the raw path, because the `endpoint` field is how log consumers group traffic per agent (the README's log example shows `"endpoint":"summarize"`), and two spellings of one agent would split it into two.

What does not change: every existing route resolves exactly as before; the suffixed paths sit under the same wildcard, hence behind the same auth middleware (`SetupRoutes` wraps `agents.Post("/*")` in one group) — a suffixed request without a valid JWT is `401` like any other; request body, headers, streaming relay, body-size cap, and CORS are untouched; `GET` on any agent path, suffixed or not, stays `405` as chi answers today for a wildcard registered `POST`-only (pinned by `TestSetupRoutes`, "guidance listing absent").

Enforcement: resolution stays in the one funnel in `internal/handlers/http/chat.go` (`urlPath` → `registry.ResolveAgent`, lines 35–36) — the strip-and-retry lives there and nowhere else, so no per-agent routes are registered and no second resolution path exists.

Chancery's README documents the suffixed forms in its routes table alongside `POST /<agent-path>`.

## Prior art

- The OpenAI Responses API convention: the client owns the `/responses` segment and appends it to whatever `base_url` it is given — https://platform.openai.com/docs/api-reference/responses. Accepting the suffix is what makes an agent path a conforming `base_url`.
- Dragoman already conforms from the other side: its serve mux registers `POST /responses` bare (`dragoman/internal/transport/serve/serve.go`, line 54), so a `base_url` of dragoman's origin works with a stock SDK today. Dragoman needs no change.
- Rejected: registering explicit `/responses` routes per agent at startup — the routing table is already one wildcard, and the rule belongs in the funnel.
- Rejected: `307` redirect to the bare path — an extra round trip per SDK call, and some clients drop bodies on redirect.

## Tests

### Skeleton

The deploy stack's walking skeleton includes one stock-SDK-shaped call, `POST <origin>/llm/<agent>/responses`, through the proxy ([proxy.md](proxy.md)). Chancery's own slice of it is that same call shape against chancery directly: a request to `/<agent>/responses` streams the backend's events back.

### Contract

Riskiest first.

- Given a config with a `responses.md` agent under a directory, when `POST /<dir>/responses` arrives, then it resolves to that agent itself — the exact match, never the stripped parent.
- Given no agent `unknown`, when `POST /unknown/responses` arrives, then the status is `404` and the body names the stripped reference `unknown`, not `unknown/responses` — the status and backend-not-reached are what `TestChatRejectsUnknownAgent` pins today; the body's reference is a new pin this case adds.
- Given an agent with named models, when `POST /<agent>.<model>/responses` arrives, then the named entry's model reaches the backend body — the strip happens before the dot split in `ResolveAgent`.
- Given auth enabled, when `POST /<agent>/responses` arrives without a JWT, then it is `401` with `WWW-Authenticate: Bearer`; with a valid token it succeeds — the suffix opens no unauthenticated path.
- Given a suffixed request, when it is forwarded, then `X-Agent` carries the stripped reference, observable at the fake backend the way `TestChatOutboundIdentity` reads it.
- Given the existing routes, when the suite runs, then nothing pinned moves: `TestChatServesEveryConfiguredRoute`, `TestChatComposesAgentFields`, `TestChatForwardsBodyUntouchedApartFromModelAndInstructions`, `TestChatRejectsUnknownAgent`, `TestChatOutboundIdentity` (chat_test.go) and `TestSetupRoutes` (routes_test.go) pass unchanged.
- Given the wildcard is `POST`-only, when `GET /<agent>/responses` arrives, then it is `405` as today.

### Isolation

Chancery's existing harness covers all of it with no other component real: `chatRouterWith` wires the real middleware chain over a registry loaded from a temp config directory, and `chatBackend` is the fake Responses backend that records what it was sent (chat_test.go style, httptest throughout). The agent-named-`responses` case adds its file through the harness's `extra` map.
