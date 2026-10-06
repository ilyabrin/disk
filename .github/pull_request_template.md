## What does this change?

<!-- One or two sentences. Link the issue it closes: "Closes #12" -->

## Checklist

- [ ] `gofmt -l .` prints nothing
- [ ] `go vet ./...` passes
- [ ] `go test -race ./...` passes
- [ ] `golangci-lint` v2.13.2 reports no issues
- [ ] New behaviour has a test, and a bug fix has a test that failed before it
- [ ] Nothing exported was changed or removed, or the PR says why it must be
- [ ] Commit messages follow [Conventional Commits](https://www.conventionalcommits.org/)
- [ ] README changes are in **both** languages

## Anything reviewers should know?

<!-- Trade-offs, things you were unsure about, follow-up work -->
