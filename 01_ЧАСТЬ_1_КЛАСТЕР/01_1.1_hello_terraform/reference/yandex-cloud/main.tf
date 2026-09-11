# Provider переводит HCL-конфигурацию в запросы к API Yandex Cloud.
# Credentials не хранятся в коде: в этой демонстрации provider читает их из окружения.
provider "yandex" {
  zone = var.zone
}

# Resource — одна конкретная VM под управлением Terraform.
resource "yandex_compute_instance" "node1" {
  name        = "node1"
  platform_id = "standard-v3"
  zone        = var.zone

  resources {
    cores  = 2
    memory = 4
  }

  boot_disk {
    initialize_params { image_id = var.image_id }
  }

  network_interface {
    subnet_id = var.subnet_id
    nat       = true
  }
}
