# main.tf — требования версий, настройка provider и объявления входных параметров.
# Имя main выбрано автором: Terraform вместе читает все .tf текущего каталога.
# Диапазон Terraform описывает совместимый язык, а lock-file фиксирует
# точную сборку provider. Плановое обновление делаем через `terraform init -upgrade`,
# после changelog и проверки plan.
terraform {
  required_version = ">= 1.9.0, < 2.0.0"

  required_providers {
    yandex = {
      source  = "yandex-cloud/yandex"
      version = "0.219.0"
    }
  }
}
# Аутентификация — через переменные окружения, НЕ через переменную с токеном в коде.
#   export YC_TOKEN=$(yc iam create-token)
#   export YC_CLOUD_ID=$(yc config get cloud-id)
#   export YC_FOLDER_ID=$(yc config get folder-id)
# Здесь токен передаётся provider через окружение, а не хранится в этом файле.
# Это не делает остальные файлы или state автоматически свободными от секретов.
provider "yandex" {
  zone = var.zone
}
variable "zone" {
  type    = string
  default = "ru-central1-a"
}

variable "cluster_name" {
  type    = string
  default = "k8s-managed"
}

variable "vpc_cidr" {
  description = "Подсеть для нод кластера"
  type        = string
  # Учебный адресный план, не обязательный диапазон Kubernetes или облака.
  # /24 = 256 адресов всего; часть адресов подсети резервирует облако.
  # 10.20.0.0/24 не пересекается с self-managed 10.10.0.0/24.
  # Для связи сетей отдельно нужны маршруты и проверка остальных диапазонов.
  default     = "10.20.0.0/24"
}

# Отдельные частные пулы Pod и Service: числа 112 и 96 — значения этого примера.
# /16 = 65 536 адресов всего, не обещание такого количества работающих Pod.
# main.tf объявляет входы; kubernetes.tf передаёт их Managed Kubernetes API.
# В self-managed эти настройки получает Kubespray через group_vars.
variable "cluster_ipv4_range" {
  description = "Pod CIDR"
  type        = string
  default     = "10.112.0.0/16"
}

variable "service_ipv4_range" {
  description = "Service CIDR"
  type        = string
  default     = "10.96.0.0/16"
}

variable "k8s_version" {
  description = "Версия Kubernetes. Проверено 29.07.2026: 1.34 в канале STABLE"
  type        = string
  default     = "1.34"
}

variable "release_channel" {
  description = "RAPID / REGULAR / STABLE. ⚠️ Сменить после создания НЕЛЬЗЯ"
  type        = string
  default     = "STABLE"
}

variable "node_count" {
  type    = number
  default = 2
}

variable "enable_autoscaling" {
  description = "Переключает node group с fixed_scale на auto_scale"
  type        = bool
  default     = false
}

variable "autoscaling_min_nodes" {
  type    = number
  default = 2
}

variable "autoscaling_max_nodes" {
  type    = number
  default = 5

  validation {
    condition     = var.autoscaling_max_nodes >= var.autoscaling_min_nodes
    error_message = "autoscaling_max_nodes must be >= autoscaling_min_nodes"
  }
}

variable "allowed_nodeport_cidr" {
  description = "Доверенный CIDR для NodePort/ICMP, обычно текущий public IP /32"
  type        = string

  validation {
    condition     = can(cidrhost(var.allowed_nodeport_cidr, 0)) && var.allowed_nodeport_cidr != "0.0.0.0/0" && var.allowed_nodeport_cidr != "::/0"
    error_message = "Укажи ограниченный CIDR, например 203.0.113.42/32."
  }
}

variable "allowed_api_cidr" {
  description = "Доверенный CIDR управления Kubernetes API, обычно текущий public IP /32"
  type        = string

  validation {
    condition     = can(cidrhost(var.allowed_api_cidr, 0)) && var.allowed_api_cidr != "0.0.0.0/0" && var.allowed_api_cidr != "::/0"
    error_message = "Укажи ограниченный CIDR управления, например 203.0.113.42/32."
  }
}
