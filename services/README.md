# Nucleus Components

> The central hub of reusable libraries and microservices.

## What is a Component?

A **component** in Nucleus is a self-contained, encapsulated capability that provides:

- **Library** (first-class citizen) — Reusable code that can be imported by satellites.
- **Service** (optional) — A microservice exposing the library via gRPC/REST.
- **MCP Server** (embedded) — Every service contains its own internal MCP server and dedicated library service.

Components follow the **vertical slicing pattern** — everything a component needs, including its specific MCP server and logic, is encapsulated within its folder.

## Component Structure

Every component follows this structure:

```
services/{domain}/{component-name}/
├── component.yaml       # Metadata (machine-readable)
├── README.md            # Documentation (human-readable)
├── lib/                 # Library code (first-class)
├── service/             # Microservice (optional)
└── examples/            # Usage examples (optional)
```

## Templates

All components should follow the standard templates for consistency. Templates are located at:

```
services/_templates/
├── component.template.yaml  # Template for component.yaml
└── README.template.md       # Template for README.md
```

> ⚠️ **Templates are living documents** — they evolve with our needs. When updating templates, ensure existing components are aligned where practical.

### Using Templates

1. Copy templates to your new component folder
2. Rename `component.template.yaml` → `component.yaml`
3. Rename `README.template.md` → `README.md`
4. Fill in the placeholders
5. Remove sections that don't apply

## Current Components

| Domain        | Component                | Status         | Description                       |
| ------------- | ------------------------ | -------------- | --------------------------------- |
| identity      | [auth](./identity/auth/) | 🚧 Development | Authentication powered by Zitadel |
| observability | logging                  | 📋 Planned     | Centralized logging               |
| observability | metrics                  | 📋 Planned     | Metrics collection                |

## Creating a New Component

1. Create folder: `services/{domain}/{component-name}/`
2. Copy templates from `_templates/`
3. Fill in `component.yaml` with metadata
4. Write `README.md` following the template
5. Implement your library in `lib/`
6. (Optional) Add microservice in `service/`

## For Satellite Repositories

To consume a component from a satellite repository:

```go
import "nucleus/services/{domain}/{component-name}"
```

Each component's README contains specific quickstart instructions.

---

_Last updated: 2026-02-01_
