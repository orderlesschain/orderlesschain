#!/usr/bin/env bash

#echo "Deploying the nodes and the clients to the deployed OpenStack cluster ..."

source "${PROJECT_ABSOLUTE_PATH}"/env
start=$(date +%s)

pushd "${PROJECT_ABSOLUTE_PATH}"/certificates || exit

if [[ $BUILD_MODE == "local" ]]; then
  rm -rf ./keys ./certs ../configs/certs ./keys_local ./certs_local ../configs/certs_local
  mkdir ./keys ./certs ./keys_local ./certs_local ../configs/certs_local
  while IFS="" read -r component || [ -n "$component" ]; do
    if ! test -z "${component}"; then
      go run ./generate_cert.go --rsa-bits 1024 --ca --start-date "Jan 1 00:00:00 2020" --duration=100000h --host "${component}"
      mv cert.pem ./certs_local/"${component}".pem
      mv key.pem ./keys_local/"${component}".pem
    fi
  done <endpoints_local
  cp ./certs_local/* ../configs/certs_local
else
  rm -rf ./keys ./certs ../configs/certs ./keys_remote ./certs_remote ../configs/certs_remote
  mkdir ./keys ./certs ./keys_remote ./certs_remote ../configs/certs_remote
  while IFS="" read -r component || [ -n "$component" ]; do
    if ! test -z "${component}"; then
      go run ./generate_cert.go --rsa-bits 1024 --ca --start-date "Jan 1 00:00:00 2020" --duration=100000h --host "${component}"
      mv cert.pem ./certs_remote/"${component}".pem
      mv key.pem ./keys_remote/"${component}".pem
    fi
  done <endpoints_remote
  cp ./certs_remote/* ../configs/certs_remote
fi

if [[ $BUILD_MODE == "local" ]]; then
  cp ./certs_local/* ./certs
  cp ./keys_local/* ./keys
else
  cp ./certs_remote/* ./certs
  cp ./keys_remote/* ./keys
fi

rm certs.tgz
tar cfz certs.tgz ./certs/*

popd || exit

#node_darwin client_darwin
for BUILD_PATH in node_linux client_linux orderer_linux; do
  mkdir -p "${PROJECT_ABSOLUTE_PATH}"/build/"${BUILD_PATH}"/{data,configs}
  cp "${PROJECT_ABSOLUTE_PATH}"/configs/app.env.secret "${PROJECT_ABSOLUTE_PATH}"/build/"${BUILD_PATH}"/configs/app.env
  if [[ $BUILD_MODE == "local" ]]; then
    cp "${PROJECT_ABSOLUTE_PATH}"/configs/endpoints_local.yml "${PROJECT_ABSOLUTE_PATH}"/build/"${BUILD_PATH}"/configs/endpoints.yml
  else
    cp "${PROJECT_ABSOLUTE_PATH}"/configs/endpoints_remote.yml "${PROJECT_ABSOLUTE_PATH}"/build/"${BUILD_PATH}"/configs/endpoints.yml
  fi
done

if [[ $BUILD_MODE == "local" ]]; then
  pushd "${PROJECT_ABSOLUTE_PATH}"/deployment/ansible/ansible_local || exit
else
  pushd "${PROJECT_ABSOLUTE_PATH}"/deployment/ansible/ansible_remote || exit
fi


ansible-playbook "${PROJECT_ABSOLUTE_PATH}"/deployment/ansible/playbooks/deploy_orderer_docker.yml
ansible-playbook "${PROJECT_ABSOLUTE_PATH}"/deployment/ansible/playbooks/deploy_node_docker.yml
ansible-playbook "${PROJECT_ABSOLUTE_PATH}"/deployment/ansible/playbooks/deploy_client_docker.yml

popd || exit

end=$(date +%s)

echo Built and deployed in $(expr $end - $start) seconds.
