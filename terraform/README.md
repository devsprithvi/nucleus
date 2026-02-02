# Nucleus Terraform Modules

Reusable Infrastructure as Code modules for the Nucleus monorepo.

> This is a **module library**. Deployments (`live/`) exist within each application's `infrastructure/terraform/` directory.

---

## Structure

```
terraform/
├── _templates/            # Scaffolding patterns for new modules & deployments
└── modules/               # Reusable module library (organized by provider)
    ├── _schemas/          # module.yaml validation schemas
    ├── huggingface/
    └── aws/
```

---

## Architecture

### Provider-First Organization

Modules are grouped by cloud provider, then by domain/service.

### The `_base` Pattern

| Type                         | Purpose                    | Usage                   |
| ---------------------------- | -------------------------- | ----------------------- |
| `_base/`                     | Raw 1:1 resource wrapper   | Internal only           |
| Public (e.g., `docker-app/`) | Opinionated vertical slice | Application consumption |

Public modules consume `_base` via `source = "../_base"` and apply our standards.

---

## Module Requirements

Every module must contain:

| File           | Purpose                                     |
| -------------- | ------------------------------------------- |
| `module.yaml`  | Machine-readable metadata (see `_schemas/`) |
| `main.tf`      | Resource definitions                        |
| `variables.tf` | Input declarations                          |
| `outputs.tf`   | Output declarations                         |
| `versions.tf`  | Version constraints                         |
| `README.md`    | Documentation                               |

---

## Consuming Modules

From application deployments:

```hcl
module "ai_space" {
  source = "git::https://github.com/org/nucleus.git//terraform/modules/huggingface/spaces/docker-app?ref=v1.0.0"

  name = "my-app"
}
```

See `_templates/app-live/` for application deployment scaffolding.
