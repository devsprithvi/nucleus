# Hugging Face Modules

Two separate modules for Hugging Face resources:

| Module    | Path          | Description                                   |
| --------- | ------------- | --------------------------------------------- |
| Spaces    | `./spaces`    | Deploy Gradio, Streamlit, Docker, Static apps |
| Endpoints | `./endpoints` | Deploy Inference API endpoints                |

## Usage

### Space

```hcl
module "my_app" {
  source = "../modules/huggingface/spaces"

  name     = "my-app"
  sdk      = "docker"
  template = "zenml/zenml"
}
```

### Endpoint

```hcl
module "my_api" {
  source = "../modules/huggingface/endpoints"

  name             = "my-api"
  model_repository = "meta-llama/Llama-2-7b"
  accelerator      = "gpu"
  instance_type    = "nvidia-l4"
}
```
