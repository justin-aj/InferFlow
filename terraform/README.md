# Terraform Starter

Terraform owns the AWS baseline for InferFlow:

- VPC and subnets
- EKS control plane
- Managed node group
- IAM roles and policy attachments

The Week 1 MVP keeps this layout intentionally small and readable so the team can iterate on it during infrastructure work.

Validate from `terraform/environments/dev`:

```bash
terraform init -backend=false
terraform validate
```
