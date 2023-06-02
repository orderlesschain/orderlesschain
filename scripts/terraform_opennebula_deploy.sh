#!/usr/bin/env bash

source "${PROJECT_ABSOLUTE_PATH}"/env
export TF_VAR_opennebula_endpoint="$OPENNEBULA_ENDPOINT"
export TF_VAR_opennebula_flow_endpoint="$OPENNEBULA_FLOW_ENDPOINT"
export TF_VAR_opennebula_username="$OPENNEBULA_USERNAME"
export TF_VAR_opennebula_password="$OPENNEBULA_PASSWORD"

start=$(date +%s)

#TFVARS="terraform_opennebula_2_orgs_2_clients.tfvars"
#TFVARS="terraform_opennebula_4_orgs_4_clients.tfvars"
#TFVARS="terraform_opennebula_16_orgs_16_clients.tfvars"
TFVARS="terraform_opennebula_32_orgs_32_clients.tfvars"


cp "${PROJECT_ABSOLUTE_PATH}"/contractsbenchmarks/networks/"${TFVARS}" "${PROJECT_ABSOLUTE_PATH}"/deployment/terraform-opennebula/terraform.tfvars

pushd "${PROJECT_ABSOLUTE_PATH}"/deployment/terraform-opennebula/ || exit

terraform apply --auto-approve

popd || exit

end=$(date +%s)
echo Terrraform done in $(expr $end - $start) seconds.
