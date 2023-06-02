#!/usr/bin/env bash

source "${PROJECT_ABSOLUTE_PATH}"/env

if [[ $BUILD_MODE == "local" ]]; then
  pushd "${PROJECT_ABSOLUTE_PATH}"/deployment/ansible/ansible_local || exit
else
  pushd "${PROJECT_ABSOLUTE_PATH}"/deployment/ansible/ansible_remote || exit
fi

ansible all -m ping
ansible-playbook "${PROJECT_ABSOLUTE_PATH}"/deployment/ansible/playbooks/prepare_linux_build_env.yml

popd || exit
