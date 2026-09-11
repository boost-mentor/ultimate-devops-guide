resource "yandex_vpc_network" "this" {
  name = "${var.cluster_name}-net"
}

resource "yandex_vpc_subnet" "this" {
  name           = "${var.cluster_name}-subnet"
  zone           = var.zone
  network_id     = yandex_vpc_network.this.id
  v4_cidr_blocks = [var.vpc_cidr]
}

# Security group для узлов managed-кластера.
resource "yandex_vpc_security_group" "k8s" {
  name       = "${var.cluster_name}-sg"
  network_id = yandex_vpc_network.this.id

  ingress {
    description       = "Трафик внутри кластера"
    protocol          = "ANY"
    predefined_target = "self_security_group"
  }
  ingress {
    description       = "Служебный трафик балансировщиков Яндекса (health-check)"
    protocol          = "TCP"
    predefined_target = "loadbalancer_healthchecks"
    from_port         = 0
    to_port           = 65535
  }
  ingress {
    description    = "Kubernetes API HTTPS только из доверенного CIDR управления"
    protocol       = "TCP"
    port           = 443
    v4_cidr_blocks = [var.allowed_api_cidr]
  }
  ingress {
    description    = "Kubernetes API native port только из доверенного CIDR управления"
    protocol       = "TCP"
    port           = 6443
    v4_cidr_blocks = [var.allowed_api_cidr]
  }
  ingress {
    description    = "NodePort только из доверенного CIDR записи"
    protocol       = "TCP"
    from_port      = 30000
    to_port        = 32767
    v4_cidr_blocks = [var.allowed_nodeport_cidr]
  }
  ingress {
    description    = "ICMP только из доверенного CIDR записи"
    protocol       = "ICMP"
    v4_cidr_blocks = [var.allowed_nodeport_cidr]
  }
  egress {
    description    = "Наружу всё"
    protocol       = "ANY"
    v4_cidr_blocks = ["0.0.0.0/0"]
  }
}
