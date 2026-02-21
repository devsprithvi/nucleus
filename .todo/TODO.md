# Repository TODOs

- [ ] **CI/CD Schema Extraction Pipeline:** Build an automated pipeline that runs against the active KubeVela cluster, extracts the dynamic JSON schemas (e.g., `webservice`, `worker`), and automatically commits/pushes them into the `schemas/kubevela/` folder in this repository.
- [ ] **Configure Dependent Repositories:** Update the `.vscode/settings.json` in other repositories to target the raw `githubusercontent.com` URLs of these centralized schemas to enable auto-completion seamlessly without local clusters.
