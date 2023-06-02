#!/usr/bin/env bash

source "${PROJECT_ABSOLUTE_PATH}"/env

for BUILD_PATH in node_linux client_linux orderer_linux sequencer_linux; do
  rm -rf "${PROJECT_ABSOLUTE_PATH}"/build/"${BUILD_PATH}"
  mkdir -p "${PROJECT_ABSOLUTE_PATH}"/build/"${BUILD_PATH}"
done

if [[ $BUILD_MODE == "local" ]]; then
  pushd "${PROJECT_ABSOLUTE_PATH}"/deployment/ansible/ansible_local || exit
else
  pushd "${PROJECT_ABSOLUTE_PATH}"/deployment/ansible/ansible_remote || exit
fi

ansible-playbook "${PROJECT_ABSOLUTE_PATH}"/deployment/ansible/playbooks/build_on_linux.yml

popd || exit
