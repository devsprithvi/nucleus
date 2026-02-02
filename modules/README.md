# Nucleus Modules

> The central hub of reusable libraries and microservices.

## What is a Module?

A **module** in Nucleus is a self-contained capability that provides:

- **Library** (first-class citizen) — Reusable code that can be imported by satellites
- **Service** (optional) — A microservice exposing the library via gRPC/REST

Modules follow the **vertical slicing pattern** — everything a module needs lives within its folder.

## Module Structure

Every module follows this structure:

```
modules/{domain}/{module-name}/
├── module.yaml          # Metadata (machine-readable)
├── README.md            # Documentation (human-readable)
├── lib/                 # Library code (first-class)
├── service/             # Microservice (optional)
└── examples/            # Usage examples (optional)
```

## Templates

All modules should follow the standard templates for consistency. Templates are located at:

```
modules/_templates/
├── module.template.yaml     # Template for module.yaml
└── README.template.md       # Template for README.md
```

> ⚠️ **Templates are living documents** — they evolve with our needs. When updating templates, ensure existing modules are aligned where practical.

### Using Templates

1. Copy templates to your new module folder
2. Rename `module.template.yaml` → `module.yaml`
3. Rename `README.template.md` → `README.md`
4. Fill in the placeholders
5. Remove sections that don't apply

## Current Modules

| Domain        | Module                   | Status         | Description                       |
| ------------- | ------------------------ | -------------- | --------------------------------- |
| identity      | [auth](./identity/auth/) | 🚧 Development | Authentication powered by Zitadel |
| observability | logging                  | 📋 Planned     | Centralized logging               |
| observability | metrics                  | 📋 Planned     | Metrics collection                |

## Creating a New Module

1. Create folder: `modules/{domain}/{module-name}/`
2. Copy templates from `_templates/`
3. Fill in `module.yaml` with metadata
4. Write `README.md` following the template
5. Implement your library in `lib/`
6. (Optional) Add microservice in `service/`

## For Satellite Repositories

To consume a module from a satellite repository:

```go
import "nucleus/modules/{domain}/{module-name}"
```

Each module's README contains specific quickstart instructions.

---

_Last updated: 2026-02-01_
