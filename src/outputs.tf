output "ses_domain_identity_arn" {
  value       = module.ses.ses_domain_identity_arn
  description = "The ARN of the SES domain identity"
}

output "domain" {
  value       = local.ses_domain
  description = "The SES domain name"
}

output "smtp_password" {
  sensitive   = true
  value       = module.ses.ses_smtp_password
  description = "The SMTP password. Only available when `ses_user_enabled` is `true`. This value is stored in Terraform state, so protect the state backend with encryption and access controls."
}

output "smtp_user" {
  value       = module.ses.access_key_id
  description = "Access key ID of the IAM user. Only available when `ses_user_enabled` is `true`"
}

output "user_name" {
  value       = module.ses.user_name
  description = "Normalized name of the IAM user. Only available when `ses_user_enabled` is `true`"
}

output "user_unique_id" {
  value       = module.ses.user_unique_id
  description = "The unique ID of the IAM user. Only available when `ses_user_enabled` is `true`"
}

output "user_arn" {
  value       = module.ses.user_arn
  description = "The ARN of the IAM user. Only available when `ses_user_enabled` is `true`"
}
