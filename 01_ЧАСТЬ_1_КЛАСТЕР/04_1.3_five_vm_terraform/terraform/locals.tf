locals {
  # Состав машин и их роли задаём один раз. Модуль создаёт VM,
  # а inventory.tf по этим же ролям собирает группы Ansible.
  nodes = {
    node1 = {
      roles        = ["control-plane", "etcd"]
      resources    = var.control_plane_resources
      api_endpoint = true # Через эту control-plane-ноду обращаемся к API с ноутбука.
    }
    node2 = { roles = ["control-plane", "etcd", "worker"], resources = var.control_plane_resources }
    node3 = { roles = ["etcd", "worker"], resources = var.worker_resources }
    lb1 = {
      roles     = ["load-balancer"]
      resources = var.load_balancer_resources
      ha        = { overlay_ip = "10.77.0.11", priority = 150 }
    }
    lb2 = {
      roles     = ["load-balancer"]
      resources = var.load_balancer_resources
      ha        = { overlay_ip = "10.77.0.12", priority = 100 }
    }
  }

  kubernetes_nodes = {
    for name, node in local.nodes : name => node
    if !contains(node.roles, "load-balancer")
  }

  # Выбор адреса API — явная настройка машины, а не первое имя по алфавиту.
  api_endpoint_nodes = [
    for name, node in local.nodes : name
    if try(node.api_endpoint, false)
  ]
}
