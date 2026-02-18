# Repository Guidelines

## Project Structure & Module Organization
This repository is a single-package Go Telegram bot (`package main`) with domain-focused files at the repo root:
- `main.go`: app bootstrap, handler registration, scheduler startup.
- `handlers.go`, `scheduler.go`, `feed*.go`, `storage.go`, `db.go`, `utils.go`: bot behavior, feed processing, persistence, helpers.
- `*_test.go`: unit tests (for storage, utils, keyboard/handlers, feed discovery).
- `Dockerfile`, `docker-compose.yaml`, `docker-entrypoint.sh`: containerized runtime and local deployment.
- `.env.example`: required environment variables template.

## Build, Test, and Development Commands
- `cp .env.example .env`: create local config; set `TG_BOT_TOKEN`.
- `docker compose up -d --build`: build and run the bot with persistent SQLite volume.
- `docker compose logs -f bot`: follow runtime logs.
- `docker compose down`: stop containers (keeps DB volume).
- `docker compose down -v`: remove containers and DB volume (destructive).
- `go test ./...`: run all unit tests.
- `go build -o feed-tgbot .`: compile local binary.
- `golangci-lint run`: lint.

## Coding Style & Naming Conventions
- Follow standard Go formatting (`gofmt`) and idioms; keep code gofmt-clean before PR.
- Use tabs/standard Go indentation (do not manually align with spaces).
- Naming: exported identifiers in `PascalCase`, internal identifiers in `camelCase`.
- Keep files grouped by behavior (`storage.go`, `handlers.go`, etc.); add new tests in matching `*_test.go`.

## Testing Guidelines
- Framework: Go standard `testing` package.
- Prefer table-driven tests with `t.Run(...)` for input/output variations.
- Keep tests deterministic and isolated (use `t.TempDir()` for temporary DB/files).
- Minimum expectation: new behavior includes tests or a clear justification for why not.

## Commit & Pull Request Guidelines
- Use Conventional Commit-style messages seen in history: `feat: ...`, `feat(scope): ...`, `chore(scope): ...`, `test: ...`.
- Keep commits focused and in imperative mood (one logical change per commit where possible).
- PRs should include a concise summary of behavior changes.
- PRs should include related issue/context.
- PRs should include test evidence (for example `go test ./...` and compose run checks).
- PRs should include screenshots or log snippets for user-visible bot interactions when relevant.

## Security & Configuration Tips
- Never commit real secrets in `.env` (especially `TG_BOT_TOKEN`).
- Rotate token immediately if exposed.
- Treat `docker compose down -v` as data-loss operation for `bot.db`.

## Important Workflow Notes
- Always run tests, linter and normalize comments before committing.
- For linter use `golangci-lint run`.
- Run tests and linter after making significant changes to verify functionality.
- Do not add comments that describe changes, progress, or historical modifications. Avoid comments like "new function," "added test," "now we changed this," or "previously used X, now using Y." Comments should only describe the current state and purpose of the code, not its history or evolution.
- After important functionality added, update README.md accordingly.
- When merging master changes to an active branch, make sure both branches are pulled and up to date first.
