# СЕРВИСНЫЕ АККАУНТЫ — самая частая точка отказа при создании managed-кластера.
# Их ДВА, и роли у них разные. Если перепутать — кластер не создастся или
# не сможет тянуть образы.

# 1. Аккаунт САМОГО КЛАСТЕРА: от его имени кластер заказывает балансировщики,
#    диски, правит security groups.
resource "yandex_iam_service_account" "cluster" {
  name        = "${var.cluster_name}-cluster-sa"
  description = "SA кластера: управляет облачными ресурсами от его имени"
}

# 2. Аккаунт УЗЛОВ: от его имени ноды скачивают образы из Container Registry.
resource "yandex_iam_service_account" "nodes" {
  name        = "${var.cluster_name}-nodes-sa"
  description = "SA узлов: тянет образы из registry"
}

data "yandex_client_config" "client" {}

# --- РОЛИ КЛАСТЕРА ---
resource "yandex_resourcemanager_folder_iam_member" "cluster_editor" {
  folder_id = data.yandex_client_config.client.folder_id
  role      = "k8s.clusters.agent" # управление ресурсами кластера
  member    = "serviceAccount:${yandex_iam_service_account.cluster.id}"
}

resource "yandex_resourcemanager_folder_iam_member" "cluster_lb" {
  folder_id = data.yandex_client_config.client.folder_id
  role      = "vpc.publicAdmin" # выдача публичных IP для LoadBalancer
  member    = "serviceAccount:${yandex_iam_service_account.cluster.id}"
}

# --- РОЛИ УЗЛОВ ---
resource "yandex_resourcemanager_folder_iam_member" "nodes_puller" {
  folder_id = data.yandex_client_config.client.folder_id
  role      = "container-registry.images.puller" # только чтение образов
  member    = "serviceAccount:${yandex_iam_service_account.nodes.id}"
}

# ⚠️ ГРАБЛЯ: роли назначаются асинхронно. Если сразу за IAM создавать кластер,
# apply может упасть с "permission denied" — роль ещё не «прожалась».
# depends_on задаёт порядок операций, но не гарантирует распространение IAM-прав.
