# Образ ОС общий для всех нод. Модуль получает уже выбранный image_id.
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-2404-lts"
}

# Cohesive typed module: root передаёт карту нод, сеть и общие настройки;
# for_each и устройство каждой ВМ инкапсулированы внутри модуля.
module "linux_vm" {
  source = "./modules/linux-vm"

  cluster_name       = var.cluster_name
  nodes              = local.nodes
  zone               = var.zone
  image_id           = data.yandex_compute_image.ubuntu.id
  subnet_id          = module.network.subnet_id
  security_group_ids = [module.network.security_group_id]
  ssh_user           = var.ssh_user
  ssh_public_key     = file(pathexpand(var.ssh_public_key_path))
}
