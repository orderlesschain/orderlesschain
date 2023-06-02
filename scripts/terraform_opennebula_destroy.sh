#!/usr/bin/env bash

source "${PROJECT_ABSOLUTE_PATH}"/env
export TF_VAR_opennebula_endpoint="$OPENNEBULA_ENDPOINT"
export TF_VAR_opennebula_flow_endpoint="$OPENNEBULA_FLOW_ENDPOINT"
export TF_VAR_opennebula_username="$OPENNEBULA_USERNAME"
export TF_VAR_opennebula_password="$OPENNEBULA_PASSWORD"

start=$(date +%s)

pushd "${PROJECT_ABSOLUTE_PATH}"/deployment/terraform-opennebula/ || exit

terraform destroy --auto-approve

popd || exit

end=$(date +%s)
echo Terrraform done in $(expr $end - $start) seconds.
