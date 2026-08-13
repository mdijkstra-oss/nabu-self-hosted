# Two stacks live in this repository. `make up` runs the container stack, the
# one Nabu self-hosted ships. `make dev` runs the same services as native
# processes that restart on a change — see README, "Development".

.DEFAULT_GOAL := help

# Checked before the stack that needs them starts, so a missing one is a
# sentence here rather than an error from inside docker or overmind.
DOCKER_TOOLS := docker
DEV_TOOLS := go npm overmind watchexec caddy

# Every port the Procfile hard-codes, and the process that wants it. The stack
# addresses these by number and nothing negotiates: a service that finds its
# port taken either dies or, in Vite's case, quietly takes the next one and
# leaves every backend rejecting it on CORS.
DEV_PORTS := 5173:frontend 8080:storage 8081:chancery 8082:embeddings 8083:dragoman

# The dev stack builds from these; the container stack fetches from GitHub and
# needs none of them present.
DEV_REPOS := ../nabu-storage ../nabu-frontend ../nabu-embeddings ../nabu-prompts ../../chancery ../../dragoman

.PHONY: help
help:
	@echo "Container stack"
	@echo "  make up        start the stack, building anything not yet built"
	@echo "  make down      stop it"
	@echo "  make rebuild   rebuild every service from its repository's main branch"
	@echo "  make logs      follow the running stack's logs"
	@echo ""
	@echo "Native stack, for development"
	@echo "  make dev       start every service as a process that restarts on a change"
	@echo "                 FORCE=1 kills whatever holds the ports first"
	@echo ""
	@echo "  make check     report what each stack needs and whether it is there"

# --- container stack ---------------------------------------------------------

.PHONY: up down rebuild logs
up: require-docker
	docker compose up

down:
	docker compose down

rebuild: require-docker
	docker compose build

logs:
	docker compose logs -f

# --- native stack ------------------------------------------------------------

# Runs in the foreground: Ctrl-C stops every process, so there is no target to
# stop it with.
#
# Closing the window instead sends SIGHUP, which stops the processes but leaves
# the socket, and overmind then refuses to start against what looks like a
# running instance. `status` answers over the socket, so its failing means
# nothing is behind it.
.PHONY: dev
dev: require-dev require-ports
	@if [ -S .overmind.sock ] && ! overmind status >/dev/null 2>&1; then \
	  echo "removing .overmind.sock, left behind by a run that did not exit"; \
	  rm -f .overmind.sock; \
	fi
	@overmind start

# --- what each stack needs ---------------------------------------------------

# One line per tool and per checkout, reporting rather than stopping: this is
# the target to run when something is wrong and you want the whole picture.
.PHONY: check
check:
	@echo "Container stack"
	@$(foreach t,$(DOCKER_TOOLS),printf "  %-11s %s\n" "$(t)" "$$(command -v $(t) || echo MISSING)";)
	@echo ""
	@echo "Native stack"
	@$(foreach t,$(DEV_TOOLS),printf "  %-11s %s\n" "$(t)" "$$(command -v $(t) || echo MISSING)";)
	@echo ""
	@echo "Checkouts the native stack builds from"
	@$(foreach r,$(DEV_REPOS),printf "  %-16s %s\n" "$(notdir $(r))" "$$([ -d $(r) ] && cd $(r) && pwd || echo 'MISSING $(r)')";)
	@echo ""
	@echo "Ports the native stack needs"
	@for entry in $(DEV_PORTS); do \
	  port=$${entry%%:*}; name=$${entry##*:}; \
	  holder=$$(lsof -nP -iTCP:$$port -sTCP:LISTEN -F c 2>/dev/null | sed -n 's/^c//p' | head -1); \
	  if [ -n "$$holder" ]; then state="TAKEN by $$holder"; else state="free"; fi; \
	  printf "  %-5s %-11s %s\n" "$$port" "$$name" "$$state"; \
	done

.PHONY: require-docker
require-docker:
	@command -v docker >/dev/null 2>&1 || { \
	  echo "docker is not installed, and the container stack is docker." >&2; \
	  echo "Install Docker Desktop, or run the native stack with 'make dev'." >&2; \
	  exit 1; }

# Names everything missing in one go. Reporting them one per run would mean a
# brew install, a rerun, and another name.
.PHONY: require-dev
require-dev:
	@missing=""; \
	for t in $(DEV_TOOLS); do \
	  command -v $$t >/dev/null 2>&1 || missing="$$missing $$t"; \
	done; \
	if [ -n "$$missing" ]; then \
	  echo "The dev stack needs these, and they are not installed:$$missing" >&2; \
	  echo >&2; \
	  echo "  brew install$$(printf '%s' "$$missing" | sed 's/ npm/ node/')" >&2; \
	  exit 1; \
	fi
	@missing=""; \
	for r in $(DEV_REPOS); do \
	  [ -d "$$r" ] || missing="$$missing $$r"; \
	done; \
	if [ -n "$$missing" ]; then \
	  echo "The dev stack builds from every Nabu repository, and these are not checked out:" >&2; \
	  for r in $$missing; do echo "  $$r" >&2; done; \
	  echo >&2; \
	  echo "README, \"Development\", has the layout it expects." >&2; \
	  exit 1; \
	fi

# Vite is the reason this is a hard stop rather than a warning. The others fail
# to bind and die loudly; Vite steps to the next free port and serves the app
# from an origin no backend has been told to allow, so the stack comes up and
# every request fails CORS.
#
# FORCE=1 kills whatever holds the ports instead of stopping: TERM first, then
# KILL for anything still listening a second later.
.PHONY: require-ports
require-ports:
	@busy=""; \
	for entry in $(DEV_PORTS); do \
	  port=$${entry%%:*}; name=$${entry##*:}; \
	  holder=$$(lsof -nP -iTCP:$$port -sTCP:LISTEN -F c 2>/dev/null | sed -n 's/^c//p' | head -1); \
	  [ -n "$$holder" ] && busy="$$busy $$port:$$name:$$holder"; \
	done; \
	[ -z "$$busy" ] && exit 0; \
	if [ -z "$(FORCE)" ]; then \
	  echo "The dev stack needs these ports, and something else is on them:" >&2; \
	  for b in $$busy; do \
	    printf "  %-5s wanted by %-10s held by %s\n" "$${b%%:*}" "$$(echo $$b | cut -d: -f2)" "$${b##*:}" >&2; \
	  done; \
	  echo >&2; \
	  echo "Stop whatever is holding them, or kill them and start anyway with:" >&2; \
	  echo "  make dev FORCE=1" >&2; \
	  exit 1; \
	fi; \
	for b in $$busy; do \
	  port=$${b%%:*}; \
	  echo "killing $${b##*:} on port $$port"; \
	  kill $$(lsof -nP -tiTCP:$$port -sTCP:LISTEN) 2>/dev/null; \
	done; \
	sleep 1; \
	for b in $$busy; do \
	  port=$${b%%:*}; \
	  pids=$$(lsof -nP -tiTCP:$$port -sTCP:LISTEN); \
	  [ -n "$$pids" ] && kill -9 $$pids 2>/dev/null; \
	done; \
	sleep 1; \
	for b in $$busy; do \
	  port=$${b%%:*}; \
	  holder=$$(lsof -nP -iTCP:$$port -sTCP:LISTEN -F c 2>/dev/null | sed -n 's/^c//p' | head -1); \
	  if [ -n "$$holder" ]; then \
	    echo "port $$port is still held by $$holder after kill -9" >&2; \
	    exit 1; \
	  fi; \
	done
