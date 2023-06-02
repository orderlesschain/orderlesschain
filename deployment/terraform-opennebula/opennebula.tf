# Variables
variable "template_id" {}
variable "group_name" {}

variable "node_vm_name" {}
variable "node_counter" {}
variable "node_cpu" {}
variable "node_memory_mb" {}
variable "node_disk_mb" {}

variable "orderer_vm_name" {}
variable "orderer_counter" {}
variable "orderer_cpu" {}
variable "orderer_memory_mb" {}
variable "orderer_disk_mb" {}

variable "sequencer_vm_name" {}
variable "sequencer_counter" {}
variable "sequencer_cpu" {}
variable "sequencer_memory_mb" {}
variable "sequencer_disk_mb" {}

variable "client_vm_name" {}
variable "client_counter" {}
variable "client_cpu" {}
variable "client_memory_mb" {}
variable "client_disk_mb" {}

variable "cli_vm_name" {}
variable "cli_counter" {}
variable "cli_cpu" {}
variable "cli_memory_mb" {}
variable "cli_disk_mb" {}

resource "opennebula_virtual_machine" "node" {
  name        = "${var.node_vm_name}-${count.index}"
  count       = var.node_counter
  cpu         = var.node_cpu
  vcpu        = var.node_cpu
  memory      = var.node_memory_mb
  template_id = var.template_id
  group       = var.group_name

  disk {
    image_id = 41
    driver   = "qcow2"
    size   = var.node_disk_mb
    target = "vda"
  }
}

resource "opennebula_virtual_machine" "orderer" {
  name        = "${var.orderer_vm_name}-${count.index}"
  count       = var.orderer_counter
  cpu         = var.orderer_cpu
  vcpu        = var.orderer_cpu
  memory      = var.orderer_memory_mb
  template_id = var.template_id
  group       = var.group_name

  disk {
    image_id = 41
    driver   = "qcow2"
    size   = var.orderer_disk_mb
    target = "vda"
  }
}

resource "opennebula_virtual_machine" "sequencer" {
  name        = "${var.sequencer_vm_name}-${count.index}"
  count       = var.sequencer_counter
  cpu         = var.sequencer_cpu
  vcpu        = var.sequencer_cpu
  memory      = var.sequencer_memory_mb
  template_id = var.template_id
  group       = var.group_name

  disk {
    image_id = 41
    driver   = "qcow2"
    size   = var.sequencer_disk_mb
    target = "vda"
  }
}

resource "opennebula_virtual_machine" "client" {
  name        = "${var.client_vm_name}-${count.index}"
  count       = var.client_counter
  cpu         = var.client_cpu
  vcpu        = var.client_cpu
  memory      = var.client_memory_mb
  template_id = var.template_id
  group       = var.group_name

  disk {
    image_id = 41
    driver   = "qcow2"
    size   = var.client_disk_mb
    target = "vda"
  }
}

resource "opennebula_virtual_machine" "cli" {
  name        = "${var.cli_vm_name}-${count.index}"
  count       = var.cli_counter
  cpu         = var.cli_cpu
  vcpu        = var.cli_cpu
  memory      = var.cli_memory_mb
  template_id = var.template_id
  group       = var.group_name

  disk {
    image_id = 41
    driver   = "qcow2"
    size   = var.cli_disk_mb
    target = "vda"
  }
}

#### The Ansible inventory file##################################################
#resource "local_file" "ansible_inventory" {
#  content  = templatefile("./templates/inventory.tmpl",
#    {
#      nodes-ips   = slice(opennebula_virtual_machine.node[*].ip, 0, 4)
#      orderer-ips = opennebula_virtual_machine.orderer[*].ip
#      sequencer-ips = opennebula_virtual_machine.sequencer[*].ip
#      clients-ips = slice(opennebula_virtual_machine.client[*].ip, 0, 4)
#      cli-ips     =  opennebula_virtual_machine.cli[*].ip
#    })
#  filename = "../ansible/ansible_remote/inventory_dir/inventory"
#}
#
#### The Nodes Endpoints
#resource "local_file" "nodes_endpoints" {
#  content  = templatefile("./templates/endpoints.tmpl",
#    {
#      nodes-ips   =  slice(opennebula_virtual_machine.node[*].ip, 0, 4)
#      orderer-ips = opennebula_virtual_machine.orderer[*].ip
#      sequencer-ips = opennebula_virtual_machine.sequencer[*].ip
#      clients-ips = slice(opennebula_virtual_machine.client[*].ip, 0, 4)
#      cli-ips     =  opennebula_virtual_machine.cli[*].ip
#    })
#  filename = "../../configs/endpoints_remote.yml"
#}
#
#### The Nodes Endpoints for creating certificates
#resource "local_file" "nodes_endpoints_certificate" {
#  content  = templatefile("./templates/endpoints_certificate.tmpl",
#    {
#      nodes-ips   =  slice(opennebula_virtual_machine.node[*].ip, 0, 4)
#      orderer-ips = opennebula_virtual_machine.orderer[*].ip
#      sequencer-ips = opennebula_virtual_machine.sequencer[*].ip
#      clients-ips = slice(opennebula_virtual_machine.client[*].ip, 0, 4)
#    })
#  filename = "../../certificates/endpoints_remote"
#}
#################################################################################

#### The Ansible inventory file##################################################
resource "local_file" "ansible_inventory" {
  content  = templatefile("./templates/inventory.tmpl",
  {
    nodes-ips   = slice(opennebula_virtual_machine.node[*].ip, 0, 16)
    orderer-ips = opennebula_virtual_machine.orderer[*].ip
    sequencer-ips = opennebula_virtual_machine.sequencer[*].ip
    clients-ips = slice(opennebula_virtual_machine.client[*].ip, 0, 32)
    cli-ips     =  opennebula_virtual_machine.cli[*].ip
  })
  filename = "../ansible/ansible_remote/inventory_dir/inventory"
}

### The Nodes Endpoints
resource "local_file" "nodes_endpoints" {
  content  = templatefile("./templates/endpoints.tmpl",
  {
    nodes-ips   =  slice(opennebula_virtual_machine.node[*].ip, 0, 16)
    orderer-ips = opennebula_virtual_machine.orderer[*].ip
    sequencer-ips = opennebula_virtual_machine.sequencer[*].ip
    clients-ips = slice(opennebula_virtual_machine.client[*].ip, 0, 32)
    cli-ips     =  opennebula_virtual_machine.cli[*].ip
  })
  filename = "../../configs/endpoints_remote.yml"
}

### The Nodes Endpoints for creating certificates
resource "local_file" "nodes_endpoints_certificate" {
  content  = templatefile("./templates/endpoints_certificate.tmpl",
  {
    nodes-ips   =  slice(opennebula_virtual_machine.node[*].ip, 0, 16)
    orderer-ips = opennebula_virtual_machine.orderer[*].ip
    sequencer-ips = opennebula_virtual_machine.sequencer[*].ip
    clients-ips = slice(opennebula_virtual_machine.client[*].ip, 0, 32)
  })
  filename = "../../certificates/endpoints_remote"
}
#################################################################################

#### The Ansible inventory file ############################################
#resource "local_file" "ansible_inventory" {
#  content  = templatefile("./templates/inventory.tmpl",
#    {
#      nodes-ips   = opennebula_virtual_machine.node[*].ip
#      orderer-ips = opennebula_virtual_machine.orderer[*].ip
#      sequencer-ips = opennebula_virtual_machine.sequencer[*].ip
#      clients-ips = opennebula_virtual_machine.client[*].ip
#      cli-ips     = opennebula_virtual_machine.cli[*].ip
#    })
#  filename = "../ansible/ansible_remote/inventory_dir/inventory"
#}
#
#### The Nodes Endpoints
#resource "local_file" "nodes_endpoints" {
#  content  = templatefile("./templates/endpoints.tmpl",
#    {
#      nodes-ips   = opennebula_virtual_machine.node[*].ip
#      orderer-ips = opennebula_virtual_machine.orderer[*].ip
#      sequencer-ips = opennebula_virtual_machine.sequencer[*].ip
#      clients-ips = opennebula_virtual_machine.client[*].ip
#      cli-ips     = opennebula_virtual_machine.cli[*].ip
#    })
#  filename = "../../configs/endpoints_remote.yml"
#}
#
#### The Nodes Endpoints for creating certificates
#resource "local_file" "nodes_endpoints_certificate" {
#  content  = templatefile("./templates/endpoints_certificate.tmpl",
#    {
#      nodes-ips   = opennebula_virtual_machine.node[*].ip
#      orderer-ips = opennebula_virtual_machine.orderer[*].ip
#      sequencer-ips = opennebula_virtual_machine.sequencer[*].ip
#      clients-ips = opennebula_virtual_machine.client[*].ip
#    })
#  filename = "../../certificates/endpoints_remote"
#}
#################################################################################

# Output
output "public_ip_nodes" {
  value = opennebula_virtual_machine.node[*].ip
}

output "public_ip_orderer" {
  value = opennebula_virtual_machine.orderer[*].ip
}

output "public_ip_sequencer" {
  value = opennebula_virtual_machine.sequencer[*].ip
}

output "public_ip_clients" {
  value = opennebula_virtual_machine.client[*].ip
}

output "public_ip_cli" {
  value = opennebula_virtual_machine.cli[*].ip
}
