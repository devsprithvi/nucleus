# .tools

This is the repository's tooling folder. It manages everything that keeps the codebase healthy, secure, and consistent — both locally and in CI/CD pipelines.

Tools are managed via `go.mod` and invoked on-demand through the root `Taskfile.yaml`. Configs live in `config/`.

## What belongs here

- **Linting** — static analysis for code, Dockerfiles, IaC, etc.
- **Secret scanning** — catching leaked credentials before they hit remote (e.g. Gitleaks, TruffleHog)
- **Vulnerability scanning** — detecting CVEs in dependencies and container images (e.g. Trivy)
- **Quality gates** — enforcing code quality standards and thresholds (e.g. SonarQube)
- **Repository hygiene** — automated dependency updates, stale branch cleanup (e.g. Renovate)
- **Anything else** that serves the repo as a whole rather than a specific application
