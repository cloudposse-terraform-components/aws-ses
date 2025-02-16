module "dns_gbl_delegated" {
  source  = "cloudposse/stack-config/yaml//modules/remote-state"
  version = "1.8.0"

  component   = "dns-delegated"
  environment = coalesce(var.dns_delegated_environment_name, module.iam_roles.global_environment_name)

  context = module.this.context
}
