terraform {
  required_version = ">= 1.0.0"

  required_providers {
    huggingface-spaces = {
      source  = "registry.terraform.io/strickvl/huggingface-spaces"
      version = "0.0.4"
    }
  }
}
