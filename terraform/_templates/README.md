# Terraform Templates

## App Infrastructure Pattern

For applications in `apps/`, we use the **Root Module** pattern.

```
apps/{app-name}/
└── terraform/
    └── live/                  # All deployment logic lives here
        ├── _base/             # ROOT MODULE: The single source of truth
        │   ├── main.tf        # Defines the app's complete infrastructure
        │   ├── variables.tf
        │   └── outputs.tf
        │
        ├── _global/           # Resources shared across environments (IAM, ECR)
        │
        ├── dev/               # Environment Instantiation
        │   └── main.tf        # source = "../_base"
        ├── staging/
        └── prod/
```

### How it works

1.  **`_base/` (The Root)**:
    You define your application's infrastructure **once** here. It orchestrates the Nucleus modules (e.g., calling `huggingface/spaces/docker-app`).

2.  **Environments (`dev/`, `prod/`)**:
    These are thin wrappers that simply feed variables to the `_base` module.

### Example

**`apps/my-app/terraform/live/_base/main.tf`** (The Blueprint)

```hcl
variable "environment" {}

# Calls the Nucleus shared module
module "space" {
  source = "git::https://github.com/org/nucleus.git//terraform/modules/huggingface/spaces/docker-app?ref=v1.0.0"

  name        = "my-app"
  environment = var.environment
  # ... other shared config
}
```

**`apps/my-app/terraform/live/dev/main.tf`** (The Deployment)

```hcl
module "root" {
  source = "../_base"

  environment = "dev"
  # dev-specific settings
}
```
