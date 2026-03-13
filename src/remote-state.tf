locals {
  dns_delegated_environment_name = var.account_map_enabled ? module.iam_roles.global_environment_name : module.this.context.environment
}

module "dns_gbl_delegated" {
  source  = "cloudposse/stack-config/yaml//modules/remote-state"
  version = "2.0.0"

  component   = "dns-delegated"
  environment = coalesce(var.dns_delegated_environment_name, local.dns_delegated_environment_name)

  bypass        = !local.enabled || var.zone_id != null
  ignore_errors = !local.enabled

  defaults = {
    default_dns_zone_id = ""
  }

  context = module.this.context
}
