#!/usr/bin/env bash
set -euo pipefail

echo "InferFlow cluster teardown"
echo "This destroys the Terraform-managed AWS baseline for the dev environment."

cd terraform/environments/dev
terraform init
terraform destroy
