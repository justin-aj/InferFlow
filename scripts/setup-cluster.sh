#!/usr/bin/env bash
set -euo pipefail

echo "InferFlow cluster bootstrap"
echo "This Week 1 script provisions AWS baseline infrastructure with Terraform and leaves in-cluster deployment to Helm/kubectl."
echo "Before running, export AWS credentials and review terraform/environments/dev/terraform.tfvars."

cd terraform/environments/dev
terraform init
terraform apply

echo "Next steps after Terraform apply:"
echo "1. Update kubeconfig for the created EKS cluster."
echo "2. Deploy the router manifest from k8s/router.yaml."
echo "3. Deploy placeholder observability/Triton assets once those components are implemented."
