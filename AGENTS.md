# Nais Administrator CLI

Guidelines for AI agents working on `narc`, the Nais Administrator CLI.

## Project Structure

- The CLI uses naistrix, which wraps Cobra. Commands are organized by domain and use typed flags. Keep command wiring and help together, with substantial operation logic in the domain packages.
- Just-in-time access grants use Google Cloud Privileged Access Manager and share Google Cloud authentication and entitlement logic.
- Kubeconfig commands build Kubernetes client configuration. Loki commands manage log deletion requests through a Kubernetes port-forward. Tenant selection uses the local Naisdevice agent.

## Common Commands

| Task | Command |
| --- | --- |
| Install development tools | `mise install` |
| Build the CLI | `mise run build` |
| Run tests with race detection and coverage | `mise run test` |
| Apply Go's suggested fixes | `mise run gofix` |
| Format Go code with gofumpt | `mise run fmt` |
| Fail if Go fixes or formatting change tracked files | `mise run ci:fmt` |
| Run all static checks | `mise run check` |
| Run staticcheck alone | `mise run check:staticcheck` |
| Generate shell completions | `mise run completions` |

## Running the CLI Locally

Build with `mise run build`. With the resulting executable on your `PATH`, use `narc --version`, `narc --help`, or `narc loki delete --help` for a smoke check.

To try shell completion in zsh, run `source <(narc completion zsh)`. The completion task generates scripts for Bash, Zsh, Fish, and PowerShell.

## Workflow and Conventions

1. Read the relevant command, flags, and domain code before editing. Follow existing patterns for naistrix registration, validation, and completion.
2. Keep help text and examples in the Go command definitions and flag tags. Check `narc <command> --help` after changing the command interface.
3. After Go changes, format with gofumpt and run `mise run test`. Run the relevant static checks, or `mise run check` for a full pass.

- Use Go's standard `testing` package and plain Go checks with `t.Errorf` or `t.Fatalf`. Avoid assertion libraries such as Testify's `assert` and `require`.
- Make all commits on a task branch; never commit directly to the default branch. Name the branch for the type and purpose of the change, such as `feat/add-tenant-filter`, `fix/loki-confirmation`, or `docs/agent-guidelines`. Use Conventional Commits for commit messages.
- Keep commit subjects to 50 characters or fewer and wrap commit body lines at 72 characters.
- Use fixtures or stubs for automated tests that touch Google Cloud, Kubernetes, or Naisdevice. Use an isolated Kubernetes configuration for kubeconfig smoke tests. `jita grant`, `jita revoke`, `tenant set`, and `loki delete` change state; use `--help` or isolated test fixtures to verify their CLI wiring.
