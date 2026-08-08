# preflight

A one-shot init service that proves every API key the stack will need exists before any other service starts; which models yaml it receives and how the rest of the stack gates on it is owned by [compose.md](compose.md).

## Contract

The check reads three inputs and nothing else: the resolved models yaml the stack hands chancery, a dragoman service-table yaml matching the table dragoman will run on, and the process environment; how each file reaches the check is [compose.md](compose.md)'s wiring.

From the models yaml it reads only two things — `models:`, the top-level map of alias name to entry, and each entry's optional `model:` string — every other field (`extends:`, `prompt:`, `reasoning_effort:`, and the rest chancery accepts) is ignored, so the check can never be stricter about those fields than chancery itself.

An entry without a literal `model:` contributes no prefix: aliases built on `extends:` inherit their model from the entry they extend, and that entry's own literal `model:` is already counted.

From dragoman.yaml it reads the top-level map of service name to entry — there is no wrapper key; dragoman's parser unmarshals the document root directly as the service map — and within each entry two fields: `auth:`, the name of the environment variable holding that service's key, and `endpoint:`, whose scheme decides whether the service can serve in this stack at all.

The derivation is mechanical: for every literal `model:` value, take the prefix before the first `/` (dragoman routes on exactly that cut, so `anthropic/claude-opus-5` addresses service `anthropic`), look the prefix up in the service map, and require the environment variable its `auth:` names; a matched `http(s)` service whose `auth:` is empty requires no variable.

A matched service whose `endpoint:` uses the `exec:` scheme (dragoman's spawned-CLI case) fails the check with a distinct message naming the service: this stack runs dragoman in a container holding one static binary, so a spawned-CLI service can never serve, and that mistake belongs at boot like a missing key.

`OPENAI_API_KEY` is always required regardless of the models yaml, because nabu-embeddings proxies OpenAI's embeddings API directly and injects that variable into every upstream request unconditionally.

A required variable counts as missing when it is unset or empty — compose interpolation defaults like `${OPENAI_API_KEY:-}` deliver empty strings, and an empty key fails at the provider exactly like an absent one.

On success the check exits 0 and prints nothing.

On missing keys it exits 1 and writes one stderr line per missing variable naming the variable and what demanded it — the model prefix, or the embeddings service for `OPENAI_API_KEY` — and it reports every missing variable in one run rather than stopping at the first.

On a prefix with no service — a typo'd or unknown provider, including a `model:` value containing no `/` at all, which dragoman cannot route — it exits 1 with a distinct message naming the prefix and the model value that carried it, since that mistake would otherwise surface as a 404 at first request.

On a models yaml or dragoman.yaml that fails to parse it exits 1 with a distinct message naming the file and the parse error — bring-your-own models yaml is user input, so both documents go through a real yaml parser into a typed shape, never through text matching.

The check validates presence, not validity: it makes no network calls, so a present-but-wrong key still fails at first request — the trade-off buys an offline, sub-second boot gate.

Side effects, exhaustively: reads the two provided files, reads the environment, writes stderr on failure; no network, no writes, no mutation of anything it checks.

## Prior art

dragoman parses the same service table in `dragoman/internal/config/config.go` (`config.Parse`), and its own comment states the gap this check fills — the auth variable is required but its value is not, so an unset key fails only when the service is called; the package lives under `internal/` and is not importable outside dragoman's module, so the check re-parses the two fields it needs.

chancery's `validate` subcommand (`chancery/internal/cli/validate.go`) already validates the prompts directory including models.yaml — alias resolution, extends chains and cycles — and the preflight complements it rather than replacing it: chancery knows nothing of dragoman.yaml or which environment variables keys live in.

chancery's models.yaml parser (`chancery/internal/prompts/manifest.go` with the field set in `types.go`) is the authority on that format; the check reads a strict subset of what chancery reads.

The compose init-container pattern — a short-lived service every other service gates on via `depends_on` with `service_completed_successfully` — is Docker's documented long-running-application startup ordering mechanism; the wiring lives in [compose.md](compose.md).

Rejected: a shell script grepping the yamls — parse, don't grep; hand-rolled text matching misreads comments, quoting, and nesting.

Rejected: validating inside chancery — chancery must stay ignorant of dragoman's auth wiring; the knowledge of which variable backs which provider belongs to the deploy repo.

Rejected: validating inside dragoman at boot — dragoman serves every configured service and cannot know which subset the models yaml will actually call, so it would either over-demand every key or stay silent.

Rejected: verifying key values with a provider round-trip — turns boot into a network operation with provider-dependent latency and failure modes, for a mistake first request catches anyway.

## Tests

Skeleton: with only `OPENAI_API_KEY` set and `MODELS=openai`, the check exits 0 silently and the dependent services start.

Contract, riskiest first:

- Given a custom-mounted models yaml that is not valid yaml, when the check runs, then it exits non-zero and stderr names that file and the parse problem — not a missing-key message.
- Given a models yaml with `model: acme/some-model` and a dragoman.yaml with no `acme` service, when the check runs with every key set, then it exits non-zero and stderr names the unknown prefix `acme` and the model value that used it.
- Given a models yaml whose `model:` value contains no `/`, when the check runs, then it fails the same unknown-prefix way, since dragoman cannot route the value.
- Given a models yaml with `model: claude-cli/opus`, when the check runs, then it exits non-zero naming `claude-cli` as an `exec:`-scheme service this stack cannot serve — never a missing-key message, and never a pass.
- Given the multi preset requiring `GEMINI_API_KEY` and `ANTHROPIC_API_KEY`, when only `ANTHROPIC_API_KEY` and `OPENAI_API_KEY` are set, then the check exits non-zero and stderr names `GEMINI_API_KEY` and the `gemini` prefix that demanded it.
- Given `MODELS=anthropic` with `ANTHROPIC_API_KEY` set but `OPENAI_API_KEY` unset, when the check runs, then it exits non-zero naming `OPENAI_API_KEY` and the embeddings service as its reason.
- Given several required variables missing at once, when the check runs, then stderr names all of them in the one run.
- Given a models yaml where one alias carries `model: anthropic/claude-opus-5` and two aliases only `extends:` it, when the check runs, then only `ANTHROPIC_API_KEY` (plus the always-required `OPENAI_API_KEY`) is demanded — the aliases contribute nothing.
- Given a required variable set to the empty string, when the check runs, then it counts as missing and the failure message names it.
- Given all derived variables and `OPENAI_API_KEY` set to any non-empty value — including values no provider would accept — when the check runs with no network available, then it exits 0 with no output, pinning that values are never verified against providers.

Isolation: the check's source lives in the deploy repository's `preflight/` directory ([compose.md](compose.md) owns how it is built and fed), so it runs alone as a plain executable against fixture models and dragoman yamls and a fully controlled environment — no compose, no network, no other services — and every case above is a file-pair-plus-env-vars table.
