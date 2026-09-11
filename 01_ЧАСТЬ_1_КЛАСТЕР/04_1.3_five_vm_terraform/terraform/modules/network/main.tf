terraform {
  required_providers {
    yandex = {
      source = "yandex-cloud/yandex"
    }
  }
}

resource "yandex_vpc_network" "this" {
  name = "${var.name_prefix}-net"
}

resource "yandex_vpc_subnet" "this" {
  name           = "${var.name_prefix}-subnet"
  zone           = var.zone
  network_id     = yandex_vpc_network.this.id
  v4_cidr_blocks = [var.vpc_cidr]
}

resource "yandex_vpc_security_group" "this" {
  name       = "${var.name_prefix}-sg"
  network_id = yandex_vpc_network.this.id

  ingress {
    description    = "SSH from trusted operator CIDR"
    protocol       = "TCP"
    port           = 22
    v4_cidr_blocks = var.trusted_admin_cidrs
  }

  ingress {
    description    = "Kubernetes API from trusted operator CIDR"
    protocol       = "TCP"
    port           = 6443
    v4_cidr_blocks = var.trusted_admin_cidrs
  }

  ingress {
    description       = "All node-to-node cluster traffic"
    protocol          = "ANY"
    predefined_target = "self_security_group"
  }

  ingress {
    description    = "NodePort from trusted operator CIDR"
    protocol       = "TCP"
    from_port      = 30000
    to_port        = 32767
    v4_cidr_blocks = var.trusted_admin_cidrs
  }

  egress {
    description    = "Package repositories and image registries"
    protocol       = "ANY"
    v4_cidr_blocks = ["0.0.0.0/0"]
  }
}
