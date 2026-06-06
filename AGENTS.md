# Repository Guidelines

## Agents Guidelines

- Dont directly write to file before you edit any files ask my permission first
- Provide me with your code changes when asking my permission

## Project Structure & Module Organization

- `cmd/main.go` is the application entrypoint and wires the router, middleware, and database.
- `config/` holds startup configuration such as environment loading and database bootstrapping.
- `internal/` contains the app logic:
  - `handler/` for HTTP handlers
  - `service/` for business logic
  - `storage/` for SQLite access
  - `dto/`, `model/`, and `view/` for request types, domain models, and templated pages
- `lib/` contains shared helpers such as validation, sessions, bcrypt, and SQLite error mapping.
- `public/css/` contains the Tailwind source and generated CSS. `tmp/` is for local build artifacts.

## Build, Test, and Development Commands

- `go run ./cmd/main.go` starts the server directly.
- dont need to run go build
- `templ generate` regenerates Go files from `.templ` templates in `internal/view/`.
- `./tailwindcss -i ./public/css/input.css -o ./public/css/output.css` rebuilds the stylesheet.
- `.air.toml` is set up for local watch builds with `go build -o ./tmp/main ./cmd/main.go`.

## Coding Style & Naming Conventions

- Follow standard Go formatting: tabs for indentation, `gofmt` for formatting, and short, clear identifiers.
- Use package names that match their folder names (`handler`, `service`, `storage`).
- Exported identifiers use `CamelCase`; unexported helpers use `camelCase`.
- Keep HTTP handlers thin and push reusable logic into `service/` or `lib/`.
- Dont ever write Javascript use Alpine js instead
- If html tag or component or piece of ui can be used on another page make react style component and save it on its own component folder

## Testing Guidelines

- Don't need to make any test

## Commit & Pull Request Guidelines

- Don't need to commit or push to cloud

## Configuration Notes

- Runtime values come from `.env` via `config/env.go`; at minimum, set `PORT` and `SESSION_SECRET`.
- The app uses a local SQLite database through `internal/storage`, so keep generated database files and other local-only artifacts out of commits unless intentionally changing fixtures.
