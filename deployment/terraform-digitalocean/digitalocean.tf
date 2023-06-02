# Variables
variable "ssh_key" {}

variable "image_name" {}
variable "region" {}

variable "node_vm_name" {}
variable "node_counter" {}
variable "node_size" {}

variable "orderer_vm_name" {}
variable "orderer_counter" {}
variable "orderer_size" {}

variable "client_vm_name" {}
variable "client_counter" {}
variable "client_size" {}

variable "cli_vm_name" {}
variable "cli_counter" {}
variable "cli_size" {}


resource "digitalocean_droplet" "node" {
  name     = "${var.node_vm_name}-${count.index}"
  count    = var.node_counter
  image    = var.image_name
  region   = var.region
  size     = var.node_size
  ssh_keys = [
    data.digitalocean_ssh_key.terraform.id
  ]
  connection {
    host        = self.ipv4_address
    user        = "root"
    type        = "ssh"
    private_key = file(var.pvt_key)
    timeout     = "2m"
  }
  provisioner "remote-exec" {
    inline = [
      "sudo useradd -g users -G sudo -s /bin/bash -m -c 'Ubuntu' ubuntu && echo 'ubuntu ALL=(ALL) NOPASSWD:ALL' | sudo tee -a /etc/sudoers && sudo groupadd ubuntu && sudo usermod -a -G ubuntu ubuntu",
      "mkdir -p /home/ubuntu/.ssh && echo '${var.ssh_key}' >> /home/ubuntu/.ssh/authorized_keys && sudo chown -R ubuntu  /home/ubuntu/.ssh && chmod 700  /home/ubuntu/.ssh && chmod 600  /home/ubuntu/.ssh/authorized_keys "
    ]
  }
}

resource "digitalocean_droplet" "orderer" {
  name     = "${var.orderer_vm_name}-${count.index}"
  count    = var.orderer_counter
  image    = var.image_name
  region   = var.region
  size     = var.orderer_size
  ssh_keys = [
    data.digitalocean_ssh_key.terraform.id
  ]
  connection {
    host        = self.ipv4_address
    user        = "root"
    type        = "ssh"
    private_key = file(var.pvt_key)
    timeout     = "2m"
  }
  provisioner "remote-exec" {
    inline = [
      "sudo useradd -g users -G sudo -s /bin/bash -m -c 'Ubuntu' ubuntu && echo 'ubuntu ALL=(ALL) NOPASSWD:ALL' | sudo tee -a /etc/sudoers && sudo groupadd ubuntu && sudo usermod -a -G ubuntu ubuntu",
      "mkdir -p /home/ubuntu/.ssh && echo '${var.ssh_key}' >> /home/ubuntu/.ssh/authorized_keys && sudo chown -R ubuntu  /home/ubuntu/.ssh && chmod 700  /home/ubuntu/.ssh && chmod 600  /home/ubuntu/.ssh/authorized_keys "
    ]
  }
}

resource "digitalocean_droplet" "client" {
  name     = "${var.client_vm_name}-${count.index}"
  count    = var.client_counter
  image    = var.image_name
  region   = var.region
  size     = var.client_size
  ssh_keys = [
    data.digitalocean_ssh_key.terraform.id
  ]
  connection {
    host        = self.ipv4_address
    user        = "root"
    type        = "ssh"
    private_key = file(var.pvt_key)
    timeout     = "2m"
  }
  provisioner "remote-exec" {
    inline = [
      "sudo useradd -g users -G sudo -s /bin/bash -m -c 'Ubuntu' ubuntu && echo 'ubuntu ALL=(ALL) NOPASSWD:ALL' | sudo tee -a /etc/sudoers && sudo groupadd ubuntu && sudo usermod -a -G ubuntu ubuntu",
      "mkdir -p /home/ubuntu/.ssh && echo '${var.ssh_key}' >> /home/ubuntu/.ssh/authorized_keys && sudo chown -R ubuntu  /home/ubuntu/.ssh && chmod 700  /home/ubuntu/.ssh && chmod 600  /home/ubuntu/.ssh/authorized_keys "
    ]
  }
}

resource "digitalocean_droplet" "cli" {
  name     = "${var.cli_vm_name}-${count.index}"
  count    = var.cli_counter
  image    = var.image_name
  region   = var.region
  size     = var.cli_size
  ssh_keys = [
    data.digitalocean_ssh_key.terraform.id
  ]
  connection {
    host        = self.ipv4_address
    user        = "root"
    type        = "ssh"
    private_key = file(var.pvt_key)
    timeout     = "2m"
  }
  provisioner "remote-exec" {
    inline = [
      "sudo useradd -g users -G sudo -s /bin/bash -m -c 'Ubuntu' ubuntu && echo 'ubuntu ALL=(ALL) NOPASSWD:ALL' | sudo tee -a /etc/sudoers && sudo groupadd ubuntu && sudo usermod -a -G ubuntu ubuntu",
      "mkdir -p /home/ubuntu/.ssh && echo '${var.ssh_key}' >> /home/ubuntu/.ssh/authorized_keys && sudo chown -R ubuntu  /home/ubuntu/.ssh && chmod 700  /home/ubuntu/.ssh && chmod 600  /home/ubuntu/.ssh/authorized_keys "
    ]
  }
}

#### The Ansible inventory file ############################################
resource "local_file" "ansible_inventory" {
  content = templatefile("../terraform/templates/inventory.tmpl",
    {
      nodes-ips   = digitalocean_droplet.node[*].ipv4_address
      orderer-ips = digitalocean_droplet.orderer[*].ipv4_address
      clients-ips = digitalocean_droplet.client[*].ipv4_address
      cli-ips     = digitalocean_droplet.cli[*].ipv4_address
    })
  filename = "../ansible/ansible_remote/inventory_dir/inventory"
}

### The Nodes Endpoints
resource "local_file" "nodes_endpoints" {
  content = templatefile("../terraform/templates/endpoints.tmpl",
    {
      nodes-ips   = digitalocean_droplet.node[*].ipv4_address
      orderer-ips = digitalocean_droplet.orderer[*].ipv4_address
      clients-ips = digitalocean_droplet.client[*].ipv4_address
      cli-ips     = digitalocean_droplet.cli[*].ipv4_address
    })
  filename = "../../configs/endpoints_remote.yml"
}

### The Nodes Endpoints for creating certificates
resource "local_file" "nodes_endpoints_certificate" {
  content = templatefile("../terraform/templates/endpoints_certificate.tmpl",
    {
      nodes-ips   = digitalocean_droplet.node[*].ipv4_address
      orderer-ips = digitalocean_droplet.orderer[*].ipv4_address
      clients-ips = digitalocean_droplet.client[*].ipv4_address
    })
  filename = "../../certificates/endpoints_remote"
}

### The Nodes Endpoints for opening log terminal
resource "local_file" "nodes_endpoints_terminal" {
  content = templatefile("../terraform/templates/endpoints_terminals_nodes.tmpl",
    {
      nodes-ips = slice(digitalocean_droplet.node[*].ipv4_address, 0, 1)
    })
  filename = "../../scripts/open_terminals/remote_endpoints_nodes"
}

### The Orderer Endpoints for opening log terminal
resource "local_file" "orderer_endpoints_terminal" {
  content = templatefile("../terraform/templates/endpoints_terminals_orderer.tmpl",
    {
      orderer-ips = slice(digitalocean_droplet.orderer[*].ipv4_address, 0, 1)
    })
  filename = "../../scripts/open_terminals/remote_endpoints_orderer"
}

### The Client Endpoints for opening log terminal
resource "local_file" "clients_endpoints_terminal" {
  content = templatefile("../terraform/templates/endpoints_terminals_clients.tmpl",
    {
      clients-ips = slice(digitalocean_droplet.client[*].ipv4_address, 0, 1)
    })
  filename = "../../scripts/open_terminals/remote_endpoints_clients"
}

# Output
output "public_ip_nodes" {
  value = digitalocean_droplet.node[*].ipv4_address
}

output "public_ip_orderer" {
  value = digitalocean_droplet.orderer[*].ipv4_address
}

output "public_ip_clients" {
  value = digitalocean_droplet.client[*].ipv4_address
}

output "public_ip_cli" {
  value = digitalocean_droplet.cli[*].ipv4_address
}
