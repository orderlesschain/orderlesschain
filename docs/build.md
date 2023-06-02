# Building the System

## Building

We use Makefile for building and running everything. Do a ``make help`` to see all the available commands.

## Used Technologies

1. go version go1.16.3 darwin/amd64
2. mattn/go-sqlite3 v1.14.7
3. protoc libprotoc v3.15.8
4. ansible v2.10.7
5. terraform

# Steps required on a fresh system

1. Create ssh keys for the new system created (when required for pushing to Git).
2. Clone everything: `git clone --recursive git-address`
3. Copy env.secret 'cp  env.secret env' and set at least PROJECT_ABSOLUTE_PATH
4. Install go version 1.16.3: https://tecadmin.net/how-to-install-go-on-ubuntu-20-04/
5. Install ansible 2.9.24: https://docs.w3cub.com/ansible~2.10/installation_guide/intro_installation#installing-ansible-on-ubuntu (pay attention to correct ansible version)
6. Install terraform: https://www.terraform.io/docs/cli/install/apt.html
7. If necessary install  protoc libprotoc and mattn/go-sqlite3
8. Tidy go mod to make sure all packages downloaded: `make tidy`
9. Do a `make build` to see if everything is installed alright.
10. Do a `make terraform-digitalocean-deploy` to deploy the VMs on Terraform (set the terraform config in env).
11. Do a `make prepare-remote-vms` to prepare the VMs.
12. Do a `make build-deploy-run-experiment-remote` to see if it builds and run example application correctly (before `cp app.env.secret app.env` in config).

## Protobuf on MAC
1. brew install protobuf
2. go get -u github.com/golang/protobuf/{proto,protoc-gen-go}
3. Add the go path to the environment profile when needed
