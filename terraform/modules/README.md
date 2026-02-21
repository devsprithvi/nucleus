# Module Library

> Reusable, org-level Terraform/OpenTofu modules. Consumed by live deployments in `.infra/`.

---

## Mental Model: Platform → Service Domain → Module

Modules are organized in a **layered hierarchy**. Each layer answers a different question:

| Layer              | Question                | Naming             | Example                           |
| :----------------- | :---------------------- | :----------------- | :-------------------------------- |
| **Platform**       | _Who_ do we talk to?    | `{platform-name}/` | `huggingface/`, `aws/`            |
| **Service Domain** | _What_ capability area? | `{service}/`       | `spaces/`, `compute/`, `storage/` |
| **Module**         | _How_ exactly?          | `{module-name}/`   | The leaf with `.tf` files         |

### Platform (Layer 1)

A platform is a vendor or system whose API we automate — AWS, Hugging Face, Cloudflare, GitHub, Zitadel. Each gets one folder. We deliberately avoid the word "provider" here to prevent confusion with Terraform's `provider` block.

### Service Domain (Layer 2)

A logical grouping that mirrors the platform's **own** service taxonomy. Don't invent categories — follow what the platform already calls them. For AWS: `compute/`, `storage/`, `networking/`. For Hugging Face: `spaces/`, `endpoints/`.

**Depth follows demand.** If a domain is simple enough for one module, the `.tf` files live directly in the domain folder. If it grows to need multiple opinionated modules, the domain becomes a category containing sub-modules.

### Module (Layer 3+)

The consumable unit — the folder containing `main.tf`, `variables.tf`, `outputs.tf`, and `module.yaml`. This is what live deployments in `.infra/` reference via `source`.

---

## The `.shared/` Pattern

When a service domain (or platform) needs **internal modules** that aren't meant for direct external consumption — raw resource wrappers, utility compositions — they go in `.shared/modules/` at that level.

This follows the same `.shared/` convention used in `.infra/`: **inherited context available at that level and below.** Internal modules exist to serve the public modules around them, which is exactly what "inherited context" means.

`.shared/` can exist at any level:

- **Platform level** — `{platform}/.shared/modules/` for things shared across all service domains of that platform
- **Domain level** — `{platform}/{domain}/.shared/modules/` for things shared across modules within a domain

Public modules consume shared internals via relative `source = "../.shared/modules/{name}"`.

> **Note:** `_base/` and `_global/` are **not used** in the module library. Those conventions belong exclusively to `.infra/` live deployments where they carry specific meanings (root blueprint and cross-environment resources, respectively).

---

## Module Template

Every module must contain:

| File           | Purpose                                                                |
| :------------- | :--------------------------------------------------------------------- |
| `module.yaml`  | Machine-readable metadata (validated by `_schemas/module.schema.json`) |
| `main.tf`      | Resource definitions                                                   |
| `variables.tf` | Input declarations                                                     |
| `outputs.tf`   | Output declarations                                                    |
| `versions.tf`  | Provider version constraints                                           |
| `README.md`    | Usage documentation                                                    |

Each platform folder should have its own `README.md` documenting its service domains and usage examples.

---

## Consuming Modules

From live deployments in `.infra/`:

```hcl
module "my_space" {
  source = "git::https://github.com/org/nucleus.git//terraform/modules/huggingface/spaces?ref=v1.0.0"

  name = "my-app"
}
```

See `_templates/` for scaffolding patterns.

---

## Meta Folders

| Folder      | Purpose                                         |
| :---------- | :---------------------------------------------- |
| `_schemas/` | JSON schemas for validating `module.yaml` files |

---

_Last updated: 2026-02-13_
