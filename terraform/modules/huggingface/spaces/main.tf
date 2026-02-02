resource "huggingface-spaces_space" "this" {
  name       = var.name
  private    = var.private
  sdk        = var.sdk
  template   = var.template
  hardware   = var.hardware
  storage    = var.storage
  sleep_time = var.sleep_time
  secrets    = var.secrets
  variables  = var.variables
}
