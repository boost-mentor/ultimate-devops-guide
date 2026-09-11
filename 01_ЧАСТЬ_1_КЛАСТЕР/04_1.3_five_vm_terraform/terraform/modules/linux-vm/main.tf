terraform {
  required_providers {
    yandex = {
      source = "yandex-cloud/yandex"
    }
  }
}

resource "yandex_compute_instance" "this" {
  for_each = var.nodes

  name        = "${var.cluster_name}-${each.key}"
  hostname    = each.key
  zone        = var.zone
  platform_id = "standard-v3"

  resources {
    cores  = each.value.resources.cores
    memory = each.value.resources.memory
  }

  boot_disk {
    initialize_params {
      image_id = var.image_id
      size     = each.value.resources.disk
      type     = "network-ssd"
    }
  }

  network_interface {
    subnet_id          = var.subnet_id
    nat                = true
    security_group_ids = var.security_group_ids
  }

  metadata = {
    ssh-keys = "${var.ssh_user}:${var.ssh_public_key}"
  }

  scheduling_policy {
    preemptible = false
  }
}
