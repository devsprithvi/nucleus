resource "huggingface_endpoint" "this" {
  name = var.name
  type = var.type

  cloud = {
    vendor = var.cloud_vendor
    region = var.cloud_region
  }

  compute = {
    accelerator   = var.accelerator
    instance_size = var.instance_size
    instance_type = var.instance_type

    scaling = {
      min_replica           = var.min_replicas
      max_replica           = var.max_replicas
      scale_to_zero_timeout = var.scale_to_zero_timeout
    }
  }

  model = {
    repository = var.model_repository
    task       = var.model_task
    framework  = var.model_framework
    revision   = var.model_revision

    image = {
      huggingface = var.image_huggingface
    }
  }
}
