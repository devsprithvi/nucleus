# Nucleus

## Project Structure

The codebase is organized into two main categories based on **consumption**:

- **`services/`**: Components that are **consumed** by other parts of the system.
  - *Example*: **Authentication (Zitadel)** resides here because it provides identity services consumed by other apps.
- **`apps/`**: Standalone applications that are **not** consumed by other internal services.
  - *Example*: **Coder Server** resides here because it is a standalone environment used by our developers, not consumed by other services.

> **Crucial Rule**: Any **App** must be moved to the **Service** category if it starts being consumed by other services.

## Getting Started

To install dependencies:

```bash
bun install
```

To run:

```bash
bun run index.ts
```

This project was created using `bun init` in bun v1.3.9. [Bun](https://bun.com) is a fast all-in-one JavaScript runtime.
