terraform {
  required_providers {
    opennebula = {
      source  = "OpenNebula/opennebula"
      version = "1.0.1"
    }
  }
}

variable "opennebula_endpoint" {}
variable "opennebula_flow_endpoint" {}
variable "opennebula_username" {}
variable "opennebula_password" {}


provider "opennebula" {
  endpoint      = var.opennebula_endpoint
  flow_endpoint = var.opennebula_flow_endpoint
  username      = var.opennebula_username
  password      = var.opennebula_password
  insecure      = true
}

