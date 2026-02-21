# Nucleus Infrastructure

> Live infrastructure deployments for this repository. Uses the reusable modules from `terraform/`.

---

## Architecture

We follow a **topology-first** layout: infrastructure mirrors the service/app hierarchy, with environments as leaf nodes inside each service.

```
.infra/
├── .shared/                          # Root-level shared config (recursive pattern)
│   ├── config/
│   │   └── backend.hcl              #   Scalr backend partial config
│   └── modules/                     #   Shared local module wrappers
│
├── _global/                          # Repo-wide global resources (DNS, org IAM)
├── _modules/                         # Repo-specific local module wrappers
│
├── services/                         # Mirrors services/ topology
│   └── {domain}/
│       └── {service}/
│           ├── .shared/             # (Optional) Service-level shared config
│           ├── _base/               # Root Module — the blueprint
│           │   ├── main.tf
│           │   ├── variables.tf
│           │   └── outputs.tf
│           ├── _global/             # Cross-environment resources
│           ├── dev/                 # Dev instantiation: source = "../_base"
│           └── prod/                # Prod instantiation: source = "../_base"
│
└── apps/                             # Mirrors apps/ topology (when needed)
    └── {app-name}/
        └── (same _base/_global/dev/prod pattern)
```

---

## Key Conventions

### Folder Prefixes

| Prefix         | Meaning                          | Recursive?       | Examples                          |
| :------------- | :------------------------------- | :--------------- | :-------------------------------- |
| **`.` prefix** | Meta/config (inherited context)  | ✅ Any level     | `.shared/`                        |
| **`_` prefix** | Structural pattern folders       | Within a service | `_base/`, `_global/`, `_modules/` |
| **No prefix**  | Domain groupings or environments | N/A              | `dev/`, `prod/`, `identity/`      |

### The Root Module Pattern (`_base/`)

Every service defines its infrastructure **once** in `_base/`. Environment folders (`dev/`, `prod/`) are thin wrappers:

```hcl
# dev/main.tf
module "root" {
  source      = "../_base"
  environment = "dev"
}
```

### `_global/` — Cross-Environment Resources

Resources that exist once per service, independent of any environment. Examples:

- IAM roles
- Container registries
- DNS records

### `.shared/` — Config Inheritance

The `.shared/` folder provides configuration that applies at its level and below. It can exist at:

- **Root** (`.infra/.shared/`) — backend config, org-wide variables
- **Service** (`.infra/services/identity/zitadel/.shared/`) — service-specific providers, tags

### `_modules/` — Repo-Specific Wrappers

Thin wrappers around `terraform/modules/` that apply Nucleus-specific defaults (naming, tagging). Not reusable org-wide.

---

## Relationship with `terraform/`

| Folder                 | Purpose                                      |
| :--------------------- | :------------------------------------------- |
| `terraform/modules/`   | Reusable module **library** (org-level)      |
| `terraform/providers/` | Custom provider **implementations**          |
| `.infra/`              | **Live deployments** consuming those modules |

```
terraform/modules/huggingface/spaces/   ← Reusable module
                ↑
                │ source = "git::...//terraform/modules/huggingface/spaces?ref=v1.0.0"
                │
.infra/services/identity/zitadel/_base/main.tf   ← Live deployment
```

---

## Workflow

1. `cd .infra/services/{domain}/{service}/dev/`
2. `tofu init -backend-config=../../../../.shared/config/backend.hcl`
3. `tofu plan`
4. `tofu apply`

---

_Last updated: 2026-02-12_
