output "cluster_id" {
  value     = yandex_kubernetes_cluster.this.id
  sensitive = true
}

output "cluster_endpoint" {
  description = "Внешний адрес API-сервера"
  value       = yandex_kubernetes_cluster.this.master[0].external_v4_endpoint
  sensitive   = true
}

output "k8s_version" { value = yandex_kubernetes_cluster.this.master[0].version }

output "get_credentials_command" {
  description = "Команда получения kubeconfig"
  value       = "yc managed-kubernetes cluster get-credentials --id ${yandex_kubernetes_cluster.this.id} --external --force"
  sensitive   = true
}

output "cidrs" {
  description = "Все диапазоны рядом — проверить, что не пересекаются"
  value = {
    nodes_subnet = var.vpc_cidr
    pod_cidr     = var.cluster_ipv4_range
    service_cidr = var.service_ipv4_range
  }
}
