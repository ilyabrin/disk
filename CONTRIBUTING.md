# Contributing to disk

Thanks for taking the time. Bug reports, fixes and new API coverage are all
welcome.

> Found a **security** problem? Do not open an issue. See [SECURITY.md](SECURITY.md).

By taking part you agree to follow the [Code of Conduct](CODE_OF_CONDUCT.md).

Not sure whether something is a bug, or want to talk an idea through first?
[Discussions](https://github.com/ilyabrin/disk/discussions) is the place.

## Getting set up

```sh
git clone https://github.com/ilyabrin/disk && cd disk
go test ./...
```

You need **Go 1.23+**. The tests use `httptest` servers throughout, so they run
offline and need no Yandex account or token.

To try a change against the real API, the programs in [examples/](examples/)
read a token from the environment:

```sh
YANDEX_DISK_TOKEN=<token-for-a-test-account> go run ./examples/demo
```

Use a throwaway account. Several examples create, move and delete files.

## Before you open a pull request

Please run the full set locally before pushing. Every one of these is also a CI
step, and each catches something a reviewer should not have to:

```sh
gofmt -l .                    # must print nothing
go vet ./...
go test -race -cover ./...
go mod tidy                   # must leave go.mod and go.sum unchanged
```

The linter is pinned in CI, so run the same version to see the same findings:

```sh
go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.13.2 run ./...
```

Its config is [.golangci.yml](.golangci.yml). gosec, `govulncheck`, CodeQL and
dependency review run in a separate workflow on every pull request.

## Compatibility

This is a library other people build on, so the public API follows
[semantic versioning](https://semver.org/):

- **Adding** a function, method, type or struct field is fine in a minor or
  patch release.
- **Changing or removing** anything exported is a breaking change and has to
  wait for a new major version, with a `/v2` module path. That includes
  changing a function's parameters or return types, and changing what an
  exported error or status value means.
- **Behaviour** counts too. A method that used to retry, wait, or return a
  `*Link` should keep doing so unless the old behaviour was a bug.

If a fix seems to need a breaking change, say so in the issue or PR, and we can
look for a compatible route first. `CreateDirAll` sitting next to `CreateDir`,
rather than replacing it, is an example of that.

## Writing code here

- **Every request takes a `context.Context`.** Use `http.NewRequestWithContext`,
  never `http.NewRequest`.
- **Close every response body**, including on error paths.
- **Use the constants from `net/http`**, such as `http.MethodGet` and
  `http.StatusAccepted`, rather than bare strings and numbers. The linter
  enforces this.
- **Wrap errors with `%w`** so callers can use `errors.Is` and `errors.As`.
- **Streamed transfers must not use `c.HTTPClient`.** Its `Timeout` is an
  absolute deadline over the whole request, body included, and aborts any
  upload or download slower than it. Use `c.transferClient()` and
  `transferContext()`, as `upload.go` and `download.go` do.

Exported identifiers need a doc comment that starts with their name. This shows
up on [pkg.go.dev](https://pkg.go.dev/github.com/ilyabrin/disk), which is where
most people first meet the library.

## Tests

New behaviour needs a test, and a bug fix needs a test that failed before the
fix. Tests talk to an `httptest` server rather than to Yandex; look at
`createdirall_test.go` or `transfer_timeout_test.go` for the pattern.

Keep tests deterministic. If one needs to wait, give it a timeout measured in
milliseconds, not seconds, so the suite stays fast.

### Against the real API

`integration_test.go` checks the library against Yandex itself. It is behind
a build tag, so `go test ./...` skips it. Run it with a token of an account
you can spare:

```sh
YANDEX_DISK_ACCESS_TOKEN=... go test -tags integration -run Integration -v ./
```

It takes about two minutes and needs a few megabytes. Everything it creates
lives in a new `disk:/disk-it-<date>-<time>` folder, which it deletes for
good at the end. A guard in the HTTP transport refuses any change outside
that folder, any trash item the test did not put there, and emptying the
whole trash. Calls that read the whole Disk log only counts, never names.

Run it before a release, and whenever you change how a request is built.

## Commit messages

We use [Conventional Commits](https://www.conventionalcommits.org/):

| Prefix      | Use for                                  |
| ----------- | ---------------------------------------- |
| `feat:`     | new API or new user-visible capability   |
| `fix:`      | bug fix                                  |
| `perf:`     | performance improvement                  |
| `refactor:` | internal change with no behaviour change |
| `test:`     | tests only                               |
| `docs:`     | documentation only                       |
| `ci:`       | workflows and tooling                    |
| `chore:`    | housekeeping                             |

Write the subject in the imperative: `fix: close the body when decoding fails`.
If a commit breaks compatibility, say so in the body with `BREAKING CHANGE:`.

## Documentation

The README exists in English and Russian. If you change one, please change the
other, or say in the PR that you could not and it will be handled. Code samples
in the README should compile against the current API.

Two conventions for anything you write in Markdown:

- **No em dashes.** They disrupt the layout GitHub renders and pull attention
  to themselves. A comma, a colon or a full stop reads better.
- **Lead with what the reader is trying to do**, then the detail.

Markdown is checked with `markdownlint`, configured in
[.markdownlint.json](.markdownlint.json):

```sh
npx markdownlint-cli *.md
```

## Licence

By contributing you agree that your work is licensed under the
[MIT License](LICENSE), matching the project.
