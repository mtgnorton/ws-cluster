# Repository Guidelines

## Project Structure & Module Organization
- Entry point: `main.go` (built by the Makefile); related command scaffold lives in `cmd/ws-cluster/`.
- Core runtime code lives in `internal/` (API, cluster, config, metrics, websocket) and `core/` (manager/checking logic).
- Protocol/transport layers are in `ws/` and `http/`; reusable helpers are in `shared/` and `pkg/`.
- Logging/instrumentation helpers live in `logger/` and `tools/` (swagger, prometheus, sentry).
- Deployment and runtime assets: `conf/` for YAML configs, `k8s/` for manifests, `docs/` for generated swagger, `bin/` for build output, `logs/` for runtime logs.

## Build, Test, and Development Commands
- `make build`: compile the service into `bin/ws-cluster`.
- `make build-linux`: static linux/amd64 build for deployment.
- `make run-local`: build and run with `--queue redis --config conf/config.yaml --env local`.
- `make run-dev` / `make run-prod`: start via `nohup` with `conf/config.dev.yaml` or `conf/config.prod.yaml`.
- `make tail-log`: tail `logs/normal.log`.
- `make build-docker VERSION=1.2.3`: build the Docker image (see `conf/config.docker.official.yaml`).
- `make run-docker`: run the latest image on port 8084.
- `go test ./...`: run all Go tests.

## Coding Style & Naming Conventions
- Go 1.22.4 (per `go.mod`); format with `gofmt` (tabs for indentation).
- Package names are lowercase; exported symbols use `PascalCase`, unexported use `camelCase`.
- Tests live in `*_test.go` and typically mirror package names.

## Testing Guidelines
- Tests are primarily under `shared/kit/`, `shared/auth/`, and `core/`.
- Run all tests with `go test ./...`; no explicit coverage threshold is enforced in the repo.

## Commit & Pull Request Guidelines
- Commit history follows Conventional Commits: `feat:`, `fix:`, `chore:`, `docs(scope):` (often with a short Chinese description).
- If opening a PR, include a short summary, test results (or “not run”), and note any config changes in `conf/` or deployment impacts.

## Configuration & Runtime Notes
- Runtime flags commonly include `--queue`, `--config`, and `--env` (see `make run-local`).
- Environment-specific configs live in `conf/` (e.g., `config.dev.yaml`, `config.prod.yaml`, `config.docker.official.yaml`).
- `nohup.out` captures background runs started by the Makefile; application logs land in `logs/`.
