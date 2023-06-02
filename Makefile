include env

UNAME := $(shell uname)
GO_PATH := $(shell which go)

ifeq ($(UNAME), Linux)
NODE_BUILD_PATH = ${NODE_BUILD_PATH_LINUX}
CLIENT_BUILD_PATH = ${CLIENT_BUILD_PATH_LINUX}
ORDERER_BUILD_PATH = ${ORDERER_BUILD_PATH_LINUX}
SEQUENCER_BUILD_PATH = ${SEQUENCER_BUILD_PATH_LINUX}
SOURCE = .
endif

ifeq ($(UNAME), Darwin)
NODE_BUILD_PATH = ${NODE_BUILD_PATH_DARWIN}
CLIENT_BUILD_PATH = ${CLIENT_BUILD_PATH_DARWIN}
ORDERER_BUILD_PATH = ${ORDERER_BUILD_PATH_DARWIN}
SEQUENCER_BUILD_PATH = ${SEQUENCER_BUILD_PATH_DARWIN}
SOURCE = source
endif

ifeq ($(GO_PATH),)
	GO_PATH = ${GO_REMOTE_PATH}
endif

.PHONY: help
## help: prints this help message
help:
	@echo "Usage: \n"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

.PHONY: tidy
## tidy: tidy up go modules
tidy:
	@${GO_PATH} mod tidy

# helper rule for deployment
check-environment:
ifndef OPENSTACK_PASSWORD_INPUT
    $(error environmet values not set)
endif

.PHONY: build
## build: build the node and the client
build: clean build-node build-orderer build-sequencer build-client

.PHONY: build-node
## build-node: build the node component (OS-dependent)
build-node: check-environment
	@echo "Building Node ..."
	@${GO_PATH} build -o ${NODE_BUILD_PATH}${NODE_BINARY} ./cmd/node

.PHONY: build-orderer
## build-orderer: build the orderer component (OS-dependent)
build-orderer: check-environment
	@echo "Building Orderer ..."
	@${GO_PATH} build -o ${ORDERER_BUILD_PATH}${ORDERER_BINARY} ./cmd/orderer

.PHONY: build-client
## build-client: build the client component (OS-dependent)
build-client:check-environment
	@echo "Building Client ..."
	@${GO_PATH} build -o ${CLIENT_BUILD_PATH}${CLIENT_BINARY} ./cmd/client

.PHONY: build-sequencer
## build-sequencer: build the sequencer component (OS-dependent)
build-sequencer: check-environment
	@echo "Building Sequencer ..."
	@${GO_PATH} build -o ${SEQUENCER_BUILD_PATH}${SEQUENCER_BINARY} ./cmd/sequencer

.PHONY: build-remote-linux
## build-remote-linux: build for remote Linux on Darwin
build-remote-linux: check-environment
	@echo "Building the components on the build remote system..."
	@PROJECT_ABSOLUTE_PATH=${PROJECT_ABSOLUTE_PATH} BUILD_MODE="remote" ./scripts/build_on_linux.sh

.PHONY: build-local-linux
## build-local-linux: build for local Linux on Darwin
build-local-linux: check-environment
	@echo "Building the components on the build local system..."
	@PROJECT_ABSOLUTE_PATH=${PROJECT_ABSOLUTE_PATH} BUILD_MODE="local" ./scripts/build_on_linux.sh

.PHONY: run-node
## run-node: run the node component
run-node: check-environment
	@echo "Running Node ..."
	@${GO_PATH} run ./cmd/node

.PHONY: run-orderer
## run-orderer: run the orderer component
run-orderer: check-environment
	@echo "Running Orderer ..."
	@${GO_PATH} run ./cmd/orderer

.PHONY: run-sequencer
## run-sequencer: run the sequencer component
run-sequencer: check-environment
	@echo "Running Sequencer ..."
	@${GO_PATH} run ./cmd/sequencer

Coordinator=false
Benchmark="0_sample"
.PHONY: run-client
## run-client: run the client component
run-client: check-environment
	@echo "Running Client ..."
	@${GO_PATH} run ./cmd/client -coordinator=${Coordinator} -benchmark=${Benchmark}

.PHONY: run-playground
## run-playground: run some experimental codes in ./cmd/playground
run-playground:
	@echo "Running Playground ..."
#	@${GO_PATH} run -race ./cmd/playground
	@${GO_PATH} run  ./cmd/playground

.PHONY: clean
## clean: clean for Darwin
clean: check-environment
	@echo "Cleaning"
	@${GO_PATH} clean
	@rm -rf ${NODE_BUILD_PATH}
	@rm -rf ${ORDERER_BUILD_PATH}
	@rm -rf ${CLIENT_BUILD_PATH}
	@rm -rf ${SEQUENCER_BUILD_PATH}

.PHONY: clean-all
## clean-all: clean everything
clean-all: check-environment
	@echo "Cleaning"
	@${GO_PATH} clean
	@rm -rf ${NODE_BUILD_PATH_LINUX}
	@rm -rf ${NODE_BUILD_PATH_DARWIN}
	@rm -rf ${ORDERER_BUILD_PATH_LINUX}
	@rm -rf ${ORDERER_BUILD_PATH_DARWIN}
	@rm -rf ${CLIENT_BUILD_PATH_LINUX}
	@rm -rf ${CLIENT_BUILD_PATH_DARWIN}
	@rm -rf ${SEQUENCER_BUILD_PATH_LINUX}
	@rm -rf ${SEQUENCER_BUILD_PATH_DARWIN}

.PHONY: git-commit
## git-commit: pull from master and push to master
git-commit:
	@echo "Commit"
	@git add . ; git commit -m 'auto push';

.PHONY: git-push
## git-push: push to master
git-push:
	@echo "Pushing to git master"
	@git add . ; git commit -m 'auto push'; \
		git push origin master

.PHONY: git-pull
## git-pull: pull from master
git-pull:
	@echo "Pulling from git master"
	@git pull origin master

.PHONY: git-push-archive
## git-push-archive: push to master Archive for experiments results
git-push-archive:
	@echo "Pushing to git Archive master"
	@cd orderlesschain-experiments-archive; git add . ; \
		git commit -m 'auto push'; git push origin main

.PHONY: protos
## protos: generate the protos
protos:
	@echo "Generating the Protos ..."
	@rm -rf ./protos/goprotos;
	@mkdir ./protos/goprotos
	@protoc -I ./protos ./protos/*.proto  --go_out=./protos/goprotos
	@protoc -I ./protos ./protos/*.proto  --go-grpc_out=require_unimplemented_servers=false:./protos/goprotos

.PHONY: terraform-deploy-prepare-vms
## terraform-deploy-prepare-vms: deploy the current plan of terraform and prepare remote VMs
terraform-deploy-prepare-vms: check-environment terraform-deploy sleep-shortly prepare-remote-vms
	@echo "Terraform deploying and preparing remote VMS..."

.PHONY: sleep-shortly
sleep-shortly:
	@echo "Sleeping for 90 seconds ..."
	@sleep 90

.PHONY: terraform-deploy
## terraform-deploy: deploy the current plan of terraform
terraform-deploy: check-environment
	@echo "Terraform deploying ..."
	@PROJECT_ABSOLUTE_PATH=${PROJECT_ABSOLUTE_PATH} ./scripts/terraform_opennebula_deploy.sh

.PHONY: terraform-destroy
## terraform-destroy: destroy the current plan of terraform
terraform-destroy: check-environment
	@echo "Terraform destroying ..."
	@PROJECT_ABSOLUTE_PATH=${PROJECT_ABSOLUTE_PATH} ./scripts/terraform_opennebula_destroy.sh

.PHONY: terraform-digitalocean-deploy
## terraform-digitalocean-deploy: deploy the current plan of terraform on Digitalocean
terraform-digitalocean-deploy: check-environment
	@echo "Terraform Digitalocean deploying ..."
	@PROJECT_ABSOLUTE_PATH=${PROJECT_ABSOLUTE_PATH} ./scripts/terraform_digitalocean_deploy.sh

.PHONY: terraform-digitalocean-destroy
## terraform-digitalocean-destroy: destroy the current plan of terraform on Digitalocean
terraform-digitalocean-destroy: check-environment
	@echo "Terraform Digitalocean destroying ..."
	@PROJECT_ABSOLUTE_PATH=${PROJECT_ABSOLUTE_PATH} ./scripts/terraform_digitalocean_destroy.sh

.PHONY: vagrant-deploy
## vagrant-deploy: deploy the vagrant locally
vagrant-deploy: check-environment
	@echo "Vagrant deploying locally ..."
	@cd ./deployment/vagrant/; vagrant up

.PHONY: vagrant-destroy
## vagrant-destroy: destroy the vagrant locally
vagrant-destroy: check-environment
	@echo "Vagrant destroy locally ..."
	@cd ./deployment/vagrant/; vagrant destroy -f

.PHONY: vagrant-suspend
## vagrant-suspend: suspend the vagrant locally
vagrant-suspend: check-environment
	@echo "Vagrant suspend locally ..."
	@cd ./deployment/vagrant/; vagrant suspend

.PHONY: vagrant-resume
## vagrant-resume: resume the vagrant locally
vagrant-resume: check-environment
	@echo "Vagrant resume locally ..."
	@cd ./deployment/vagrant/; vagrant resume

.PHONY: build-deploy-all-remote
## build-deploy-all-remote: build and deploy remote components using Ansible
build-deploy-all-remote: check-environment build-remote-linux
	@echo "Running the bash script for building and deploying components ..."
	@PROJECT_ABSOLUTE_PATH=${PROJECT_ABSOLUTE_PATH} BUILD_MODE="remote" WAN_DRIVE=${WAN_DRIVE_REMOTE} ./scripts/deploy_components.sh

.PHONY: build-deploy-all-local
## build-deploy-all-local: build and deploy local components using Ansible
build-deploy-all-local: check-environment build-local-linux
	@echo "Running the bash script for building and deploying components ..."
	@PROJECT_ABSOLUTE_PATH=${PROJECT_ABSOLUTE_PATH} BUILD_MODE="local" WAN_DRIVE=${WAN_DRIVE_LOCAL} ./scripts/deploy_components.sh

.PHONY: prepare-remote-linux-env
## prepare-remote-linux-env: Create the remote linux build env
prepare-remote-linux-env:
	@echo "Preparing the remote linux build env ..."
	@PROJECT_ABSOLUTE_PATH=${PROJECT_ABSOLUTE_PATH} BUILD_MODE="remote" ./scripts/prepare_linux_build_env.sh

.PHONY: prepare-remote-vms
## prepare-remote-vms: Install dependencies on all remote VMs
prepare-remote-vms: prepare-remote-linux-env
	@echo "Install dependencies on all remote VMs ..."

.PHONY: prepare-local-vms
## prepare-local-vms: Install dependencies on local VMs
prepare-local-vms:
	@echo "Preparing the local linux build env ..."
	@PROJECT_ABSOLUTE_PATH=${PROJECT_ABSOLUTE_PATH} BUILD_MODE="local" ./scripts/prepare_linux_build_env.sh

.PHONY: run-example-experiment-remote
## run-example-experiment-remote: Run the example experiment remote
run-example-experiment-remote:
	@echo "Running the example experiment remote..."
	@PROJECT_ABSOLUTE_PATH=${PROJECT_ABSOLUTE_PATH} BUILD_MODE="remote" ./scripts/run_example_experiment.sh

.PHONY: build-deploy-run-experiment-remote
## build-deploy-run-experiment-remote: Build deploy and run the example experiment remote
build-deploy-run-experiment-remote: build-deploy-all-remote run-example-experiment-remote
	@echo "Build deploy and run the example experiment remote..."

.PHONY: run-all-experiments-remote
## run-all-experiments: Run ALL experiments remote
run-all-experiments-remote:
	@echo "Running ALL experiments remote..."
	@PROJECT_ABSOLUTE_PATH=${PROJECT_ABSOLUTE_PATH} BUILD_MODE="remote" ./scripts/run_all_experiments.sh

.PHONY: run-extreme-experiments-remote
## run-extreme-experiments-remote: Run all extreme experiments remote
run-extreme-experiments-remote:
	@echo "Running all extreme experiments remote..."
	@PROJECT_ABSOLUTE_PATH=${PROJECT_ABSOLUTE_PATH} BUILD_MODE="remote" ./scripts/run_extreme_experiments.sh

.PHONY: run-example-experiment-local
## run-example-experiment-local: Run the example experiment local
run-example-experiment-local:
	@echo "Running the example experiment locally ..."
	@PROJECT_ABSOLUTE_PATH=${PROJECT_ABSOLUTE_PATH} BUILD_MODE="local" ./scripts/run_example_experiment.sh

.PHONY: build-deploy-run-experiment-local
## build-deploy-run-experiment-local: Build deploy and run the example experiment local
build-deploy-run-experiment-local: build-deploy-all-local run-example-experiment-local
	@echo " Build deploy and run the example experiment local..."

.PHONY: build-docker-images
## build-docker-images: build docker images
build-docker-images:
	@echo "Building Docker images ..."
	@docker build --force-rm -t ${DOCKER_NS_ENV}/orderlesschain-node -f ./deployment/docker/images/node.dockerfile .
	@docker build --force-rm -t ${DOCKER_NS_ENV}/orderlesschain-orderer -f ./deployment/docker/images/orderer.dockerfile .
	@docker build --force-rm -t ${DOCKER_NS_ENV}/orderlesschain-client -f ./deployment/docker/images/client.dockerfile .
	@docker tag ${DOCKER_NS_ENV}/orderlesschain-node ${DOCKER_NS_ENV}/orderlesschain-node:${DOCKER_TAG_ENV}
	@docker tag ${DOCKER_NS_ENV}/orderlesschain-orderer ${DOCKER_NS_ENV}/orderlesschain-orderer:${DOCKER_TAG_ENV}
	@docker tag ${DOCKER_NS_ENV}/orderlesschain-client ${DOCKER_NS_ENV}/orderlesschain-client:${DOCKER_TAG_ENV}

.PHONY: push-docker-images
## push-docker-images: push docker images
push-docker-images:
	@echo "Pushing Docker images ..."
	@docker login --username ${DOCKER_USERNAME_ENV} --password ${DOCKER_PASSWORD_ENV}
	@docker push ${DOCKER_NS_ENV}/orderlesschain-node:${DOCKER_TAG_ENV}
	@docker push ${DOCKER_NS_ENV}/orderlesschain-orderer:${DOCKER_TAG_ENV}
	@docker push ${DOCKER_NS_ENV}/orderlesschain-client:${DOCKER_TAG_ENV}

.PHONY: install-local-docker
## install-local-docker: Install local docker
install-local-docker:
	@echo "Installing dockers locally ..."
	@PROJECT_ABSOLUTE_PATH=${PROJECT_ABSOLUTE_PATH} BUILD_MODE="local" ./scripts/install_docker.sh

.PHONY: install-remote-docker
## install-remote-docker: Install remote docker
install-remote-docker:
	@echo "Installing dockers remote ..."
	@PROJECT_ABSOLUTE_PATH=${PROJECT_ABSOLUTE_PATH} BUILD_MODE="remote" ./scripts/install_docker.sh

.PHONY: build-deploy-all-remote-docker
## build-deploy-all-remote-docker: build and deploy remote components using Ansible and Docker
build-deploy-all-remote-docker: check-environment build-docker-images push-docker-images
	@echo "Running the bash script for building and deploying components ..."
	@PROJECT_ABSOLUTE_PATH=${PROJECT_ABSOLUTE_PATH} BUILD_MODE="remote" WAN_DRIVE=${WAN_DRIVE_REMOTE}  ./scripts/deploy_components_with_docker.sh

.PHONY: build-deploy-all-local-docker
## build-deploy-all-local-docker: build and deploy local components using Ansible and Docker
build-deploy-all-local-docker: check-environment build-docker-images push-docker-images
	@echo "Running the bash script for building and deploying components ..."
	@PROJECT_ABSOLUTE_PATH=${PROJECT_ABSOLUTE_PATH} BUILD_MODE="local" WAN_DRIVE=${WAN_DRIVE_LOCAL} ./scripts/deploy_components_with_docker.sh
