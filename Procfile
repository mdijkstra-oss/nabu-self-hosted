# Nabu running natively, one process per repository, each rebuilt and
# restarted on a change to its own source. `overmind start` from here.
#
# The port map is this file and nothing else. Inside the container stack every
# service has its own network namespace, so storage and dragoman can both hold
# 8080; sharing one host they cannot, and dragoman moves to 8083.
#
#   storage 8080   chancery 8081   embeddings 8082   dragoman 8083   app 5173
#
# API keys and PROJECT_DIR come from .env, the same file compose reads. The
# *_REPO variables override where each repository lives; unset, the layout in
# the README applies: siblings of this repo, chancery and dragoman one level up.
#
# The Go services build to .dev/ and exec the binary rather than going through
# `go run`, whose child survives the signal watchexec sends to restart it.

storage: PORT=8080 PERSISTENCE_DIR=${PROJECT_DIR:?set PROJECT_DIR in .env} CORS_ORIGINS=http://localhost:5173 watchexec -rq -e go -w "${NABU_STORAGE_REPO:-../nabu-storage}" -- 'mkdir -p .dev && go build -C "${NABU_STORAGE_REPO:-../nabu-storage}" -o "$PWD/.dev/storage" ./cmd && exec .dev/storage'

dragoman: watchexec -rq -e go,yaml -w "${DRAGOMAN_REPO:-../../dragoman}/cmd" -w "${DRAGOMAN_REPO:-../../dragoman}/internal" -- 'mkdir -p .dev && go build -C "${DRAGOMAN_REPO:-../../dragoman}" -o "$PWD/.dev/dragoman" ./cmd/dragoman && exec .dev/dragoman serve --addr 127.0.0.1:8083'

# Watches nabu-prompts too: the prompts are chancery's input, and it reads them
# once at boot, so editing one is a change to this service.
chancery: PORT=8081 RESPONSES_BASE_URL=http://localhost:8083 CORS_ORIGINS=http://localhost:5173 watchexec -rq -e go,md,yaml -w "${CHANCERY_REPO:-../../chancery}" -w "${NABU_PROMPTS_REPO:-../nabu-prompts}/config" -- 'mkdir -p .dev && go build -C "${CHANCERY_REPO:-../../chancery}" -o "$PWD/.dev/chancery" ./cmd/chancery && exec .dev/chancery serve --config "${NABU_PROMPTS_REPO:-../nabu-prompts}/config" --models "${MODELS:-models.openai.yaml}"'

# CORS_ORIGIN, not CORS_ORIGINS: this one is Caddy's and takes a single origin.
embeddings: CORS_ORIGIN=http://localhost:5173 watchexec -rq -w "${NABU_EMBEDDINGS_REPO:-../nabu-embeddings}/Caddyfile" -- caddy run --adapter caddyfile --config "${NABU_EMBEDDINGS_REPO:-../nabu-embeddings}/Caddyfile"

frontend: VITE_API_HOST=localhost:8080 VITE_LLM_HOST=http://localhost:8081 VITE_EMBEDDINGS_URL=http://localhost:8082/embeddings npm --prefix "${NABU_FRONTEND_REPO:-../nabu-frontend}" run dev
