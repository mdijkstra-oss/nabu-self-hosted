# Welcome seed

A one-shot service in the deploy stack: when storage holds zero projects, it creates one welcome project, so the app's existing first-project redirect (`app/routes/home.tsx` navigates to the first project once the list loads) lands a first boot inside a project instead of on "No projects found".

## Contract

The seed runs once storage is healthy ([compose.md](compose.md) owns the wiring) and talks to storage directly on the compose network, never through the proxy.

It reads one thing: `GET /queries/projects` on storage; a first page with zero projects is the empty signal, anything else means the seed does nothing and exits 0.

On empty it makes one write: `POST /commands/{projectId}` with a fresh UUID as the id and a JSON body of `action` `WriteFile`, `path` `welcome.md`, and a fixed embedded welcome text beginning `# Welcome to Nabu!` — uncompressed, far under storage's caps. That write is what brings the project into existence; storage's own template seeding (`internal/domain/files/seed.go`) adds the required `preferences.md` and `settings.hidden.md` at the first WebSocket connect, so the seed never duplicates them.

The empty check is projects, not files: a user who deletes `welcome.md` but keeps any project is never re-seeded; only a storage with no projects at all is seeded again on the next boot.

On failure — storage unreachable, or a non-2xx answer — it exits non-zero with the response on stderr, and the stack keeps serving: nothing gates on this service, so the worst outcome is the pre-seed behavior, an empty project list.

Side effects, exhaustively: at most two HTTP calls to storage; no disk, no environment, no other service.

Enforcement: the seed speaks only the two storage surfaces linked above, and the welcome content is a fixed string compiled into the seed, never input.

## Prior art

Storage's own `SeedRequiredFiles` (`nabu-storage/internal/domain/files/seed.go`) is the same idea one level down — template files into an existing project — and is deliberately not reusable here: it early-returns on nonexistent projects, and storage is pinned unchanged by [spec.md](spec.md).

prompts-seed in [compose.md](compose.md) is the stack's existing one-shot service idiom this follows.

Rejected: patching storage to auto-create a project at boot — what a first project contains is product content, not storage's concern, and storage stays untouched.

Rejected: waiting for a frontend create-project flow — wanted eventually (a sample project is planned), but the deploy stack should not depend on a future frontend feature, and the app's existing redirect already handles everything after the project exists.

## Tests

1. Skeleton — on a clean first boot the seed creates the welcome project, and the browser, having loaded the app, is redirected into it and renders `welcome.md`.

2. Contract, riskiest first:
   Given storage with zero projects, when the seed runs, then `GET /queries/projects` lists exactly one project and it contains `welcome.md` with the welcome text.
   Given storage with one or more projects — welcome or not — when the seed runs again, then it exits 0 and nothing changed: no second welcome project, no overwritten file.
   Given a project whose `welcome.md` the user deleted, when the stack restarts, then nothing is recreated — the empty signal is the project list.
   Given storage down or erroring, when the seed runs, then it exits non-zero and every other service serves on — no dependency points at the seed.

3. Isolation — the seed's container runs against a storage container alone (storage's own image on a scratch `PERSISTENCE_DIR`), and every case above is asserted through `GET /queries/projects` and the file content; no proxy, no other service.
