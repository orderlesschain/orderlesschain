#!/usr/bin/env bash

source "${PROJECT_ABSOLUTE_PATH}"/env
export TF_VAR_do_token="$DIGITALOCEAN_DO_TOKEN"
export TF_VAR_pvt_key="$DIGITALOCEAN_PRIVATE_KEY"
export TF_VAR_public_key="$DIGITALOCEAN_PUBLIC_KEY_NAME"
export TF_VAR_ssh_key="$DIGITALOCEAN_SSH_KEY"

start=$(date +%s)

pushd "${PROJECT_ABSOLUTE_PATH}"/deployment/terraform-digitalocean/ || exit

terraform destroy --auto-approve

popd || exit

end=$(date +%s)
echo Terrraform done in $(expr $end - $start) seconds.
