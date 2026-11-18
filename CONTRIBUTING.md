# Contributing to wa-cli

Thanks for considering a contribution! wa-cli is developed one milestone at
a time — see [ROADMAP.md](./ROADMAP.md) for the current phase and what's
planned next. That's the best place to find work that fits where the
project is right now.

## Getting started

1. Fork and clone the repo.
2. Install Go 1.25+ (the version in `go.mod`).
3. `make verify` to confirm everything works locally.
4. Pick an open issue, or a task from the current roadmap phase.

`make verify` is the check every commit should pass: it builds every
package, runs `go vet`, fails if `gofmt` would change a file, and runs
the tests with the race detector. The race detector needs cgo and a C
compiler; without one, run `make verify TESTFLAGS=`.

## Workflow

- Create a branch off `main`: `git checkout -b feature/short-description`.
- Keep commits focused; write clear commit messages (imperative mood, e.g.
  "Add chat search command").
- Run `make verify` before every commit, and `make lint` before opening
  a PR (it needs [golangci-lint](https://golangci-lint.run/welcome/install/)).
- Update `CHANGELOG.md` under "Unreleased" for user-facing changes.
- Open a PR against `main` and describe what changed and why.

## Code style

- Standard `gofmt`/`go vet` cleanliness — enforced in CI.
- `golangci-lint run` must pass (config in `.golangci.yml`).
- Prefer small, focused packages under `internal/`, mirroring the existing
  structure (one concern per package).
- Commands live in `cmd/`, one file per command or group, and register
  themselves on `rootCmd` from their own `init()`. Open a WhatsApp client
  with the helpers in `cmd/client.go`.
- Before writing a helper about WhatsApp itself (JIDs, phone numbers,
  media types, sanitizing other people's text), look in
  `internal/whatsapp`: `cmd/` and `internal/tui` call those rather than
  keep their own copies. [ARCHITECTURE.md](./ARCHITECTURE.md) has the map.
- The TUI talks to WhatsApp only through the `Backend` interface in
  `internal/tui/backend.go`.

## Tests

- New packages should ship with table-driven unit tests.
- `make test` runs `go test -race -cover ./...`; CI also tracks coverage
  on core packages (target: 80%+, see Phase 15 in the roadmap).
- TUI tests drive the real model with the fake backend in
  `internal/tui/fake_test.go`: send keys with the test harness and check
  what the backend was asked to do and what the frame shows.
  `WA_DEMO=1 go run .` opens the TUI on made-up chats.

## Reporting bugs / requesting features

Open a GitHub issue. For bugs, include your OS, Go version, and the exact
command you ran. For features, note which roadmap phase (if any) it fits
under, or propose it as a "Future (v2.0)" idea.

## Code of Conduct

Participation in this project is governed by our
[Code of Conduct](./CODE_OF_CONDUCT.md).
