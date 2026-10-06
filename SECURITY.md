# Security Policy

## Supported versions

Only the latest release receives security fixes. Please upgrade before reporting.

| Version        | Supported |
| -------------- | --------- |
| latest release | ✅        |
| older          | ❌        |

## Reporting a vulnerability

**Please do not open a public issue for security problems.**

Use GitHub's private reporting instead:
[**Report a vulnerability**](https://github.com/ilyabrin/disk/security/advisories/new)

Include what you have: the affected version, Go version, a minimal snippet that
shows the problem, and the impact you believe it has. A proof of concept helps
but is not required.

You can expect an acknowledgement within **7 days** and a status update within
**30 days**. If a fix is warranted, you will be credited in the release notes
unless you prefer otherwise.

## What is in scope

This is a client library, so the questions that matter are about what it does
with the token and the data it is handed:

- The OAuth token leaking: into logs, error messages, URLs, or anywhere other
  than the `Authorization` header sent to Yandex
- Requests reaching a host other than the configured API endpoint, for example
  by following a crafted redirect or an attacker-controlled `href`
- Path handling that lets a crafted path escape what the caller intended,
  including the traversal checks in `validatePath`
- Local file handling in uploads and downloads that reads or writes outside the
  paths the caller passed in
- Response handling that can be driven into unbounded memory use or a hang

## What is not in scope

- Vulnerabilities in Yandex.Disk or its API, which should go to
  [Yandex](https://yandex.com/bugbounty/).
- Problems that need the caller to pass an already-compromised token, or to
  point `ClientConfig.BaseURL` at a server they do not trust.
- Findings in the `examples/` directory, which is illustrative code rather than
  part of the library.

## Handling your own token

The library never stores a token on disk; that is up to your application. It
sends the token only in the `Authorization` header of requests to the API.

Debug logging masks the token by default. Headers are logged only when
`LoggerConfig.Verbose` is on, and even then `SanitizeAuth: true`, the default,
logs the Authorization header as `OAuth ***`: the scheme stays visible and no
part of the credential does. Turning `SanitizeAuth` off writes the token to the
log in full, so do that only on a machine you trust and never in production.

Keep tokens out of source control, logs and shared configuration. If one
leaks, revoke it at
[yandex.ru/id/security/applications](https://yandex.ru/id/security/applications).
