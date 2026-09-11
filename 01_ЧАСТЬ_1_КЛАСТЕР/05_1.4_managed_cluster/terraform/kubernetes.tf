resource "yandex_kubernetes_cluster" "this" {
  name       = var.cluster_name
  network_id = yandex_vpc_network.this.id

  master {
    version = var.k8s_version

    # Одна master_location: control plane размещаем в одной зоне для учебного стенда.
    # Это не защита от отказа всей зоны. Для межзонной устойчивости нужны
    # control plane в трёх зонах с соответствующими подсетями, а также
    # отдельное размещение worker-нод и реплик приложения по зонам.
    master_location {
      zone      = var.zone
      subnet_id = yandex_vpc_subnet.this.id
    }

    public_ip = true # чтобы ходить kubectl'ом снаружи

    security_group_ids = [yandex_vpc_security_group.k8s.id]

    maintenance_policy {
      auto_upgrade = false # Для учебного стенда версию меняем только явным apply.
      # В канале RAPID выключить НЕЛЬЗЯ — только REGULAR/STABLE.
    }
  }

  # Pod CIDR и Service CIDR. Здесь они задаются терраформом,
  # В self-managed этот параметр задаётся через Kubespray group_vars.
  cluster_ipv4_range = var.cluster_ipv4_range
  service_ipv4_range = var.service_ipv4_range

  service_account_id      = yandex_iam_service_account.cluster.id
  node_service_account_id = yandex_iam_service_account.nodes.id

  release_channel = var.release_channel

  depends_on = [
    yandex_resourcemanager_folder_iam_member.cluster_editor,
    yandex_resourcemanager_folder_iam_member.cluster_lb,
    yandex_resourcemanager_folder_iam_member.cluster_lb_admin,
    yandex_resourcemanager_folder_iam_member.nodes_puller,
  ]
}
resource "yandex_kubernetes_node_group" "workers" {
  cluster_id = yandex_kubernetes_cluster.this.id
  name       = "${var.cluster_name}-workers"
  version    = var.k8s_version
  # ⚠️ Правило Яндекса: версия группы узлов НЕ МОЖЕТ быть выше версии мастера,
  # и разница не более ДВУХ минорных (в ванильном k8s допускается три).

  instance_template {
    platform_id = "standard-v3"

    resources {
      cores  = 2
      memory = 4
    }

    boot_disk {
      type = "network-ssd"
      size = 30
    }

    network_interface {
      nat                = true
      subnet_ids         = [yandex_vpc_subnet.this.id]
      security_group_ids = [yandex_vpc_security_group.k8s.id]
    }

    scheduling_policy { preemptible = false }

    container_runtime { type = "containerd" }
  }

  scale_policy {
    dynamic "fixed_scale" {
      for_each = var.enable_autoscaling ? [] : [1]
      content {
        size = var.node_count
      }
    }

    dynamic "auto_scale" {
      for_each = var.enable_autoscaling ? [1] : []
      content {
        min     = var.autoscaling_min_nodes
        max     = var.autoscaling_max_nodes
        initial = var.node_count
      }
    }
  }

  allocation_policy {
    location { zone = var.zone }
  }

  maintenance_policy {
    auto_upgrade = false # для стенда — контролируем момент апгрейда сами
    auto_repair  = true
  }
}
