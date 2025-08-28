# 0003 – Dev tooling (optional lint/format)

Status: Accepted

Context
- Keeping the codebase tidy improves contributions, but we want to avoid hard dependencies/tools unless optional and lightweight.

Decision
- Provide optional Makefile targets for linting/formatting:
  - `make lint` runs `ruff check` if available; otherwise falls back to `pyflakes` if available.
  - `make format` runs `black` if available.
- Do not add these tools as runtime dependencies; contributors may install them locally.
- Add `make env-init` to create a root `.env` from `.env.example` for Docker Compose convenience.

Consequences
- Contributors get a consistent command surface without mandatory new dependencies.
- CI (if added) can choose to enforce ruff/black by installing them explicitly.
