# Infrastructure-as-Code Conventions

> [!important]
> These rules apply to all Terraform, OpenTofu, and IaC configuration. Enforce regardless of cloud provider.

---

## Naming

- Resource identifiers: `snake_case` — never repeat the resource type in the name (`aws_s3_bucket.logs`, not `aws_s3_bucket.s3_logs_bucket`)
- Cloud resource names: `kebab-case` pattern `{env}-{project}-{component}` (e.g., `prod-myapp-api`)
- Module naming: `terraform-{provider}-{purpose}` (e.g., `terraform-aws-vpc`)
- Variable and output names: `snake_case`, descriptive, no abbreviations

## File Organization

- `main.tf` — resources and data sources
- `variables.tf` — all input variables with descriptions and types
- `outputs.tf` — all outputs with descriptions
- `providers.tf` — provider configuration and required_providers
- `locals.tf` — local values and computed expressions
- `data.tf` — data sources (alternative: keep in main.tf for small configs)
- `versions.tf` — terraform version constraint

## Modules

- One module per logical component — never put the entire infrastructure in a single module
- Every variable must have `description` and `type`
- Every output must have `description`
- Pin module source versions: `source = "git::...?ref=v1.2.0"` or registry `version = "~> 3.0"`
- No hardcoded values inside modules — everything parameterized through variables

## State Management

- Always remote state: S3+DynamoDB, GCS, or Azure Blob — never local
- State locking enabled — never disable it
- Separate state per environment (`dev`, `staging`, `prod`)
- Never store secrets in state — use `sensitive = true` on variables and outputs
- Never manually edit state files — use `terraform state` commands with caution

## Tagging

- Mandatory tags on every taggable resource: `Environment`, `Project`, `Team`, `ManagedBy` ("terraform"), `CostCenter`
- Use a shared `locals` block or module for default tags — never repeat tag blocks
- Additional context tags encouraged: `Component`, `Owner`, `Repository`

## Safety

- `prevent_destroy = true` on stateful resources: databases, S3 buckets, encryption keys, DNS zones
- Never `-auto-approve` in production — always review the plan
- Always `terraform plan` before `terraform apply` — review every change
- Pin provider versions: `version = "~> 5.0"` — never leave unversioned
- Run `tfsec` or `checkov` in pre-commit hooks and CI

## Variables

- Set sensible defaults where possible — but never default secrets or environment-specific values
- Use `validation` blocks for input constraints
- Group related variables with comments
- Use `object` types for complex variable groups rather than many individual variables
