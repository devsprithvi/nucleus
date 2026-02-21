# Hugging Face Spaces Module

Simple Terraform module for creating Hugging Face Spaces.

## Usage

```hcl
module "my_space" {
  source = "./terraform/modules/huggingface/spaces"

  hf_token    = var.hf_token
  hf_username = "your-username"
  name        = "my-space"
  sdk         = "docker"
}
```

## Inputs

| Name          | Description                               | Required |
| ------------- | ----------------------------------------- | -------- |
| `hf_token`    | HF API token                              | Yes      |
| `hf_username` | HF username                               | Yes      |
| `name`        | Space name                                | Yes      |
| `sdk`         | SDK type (docker/gradio/streamlit/static) | No       |
| `private`     | Private space                             | No       |
| `secrets`     | Secret env vars                           | No       |
| `variables`   | Public env vars                           | No       |

## Outputs

| Name        | Description       |
| ----------- | ----------------- |
| `space_id`  | Full space ID     |
| `space_url` | Space page URL    |
| `embed_url` | Direct access URL |
