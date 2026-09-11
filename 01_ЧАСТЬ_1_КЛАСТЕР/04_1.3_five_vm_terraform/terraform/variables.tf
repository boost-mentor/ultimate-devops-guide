variable "zone" {
  description = "Зона доступности"
  type        = string
  default     = "ru-central1-a"
}

variable "cluster_name" {
  description = "Префикс имён ресурсов"
  type        = string
  default     = "k8s-selfmanaged"
}

# --- СЕТИ. Значения осознанные, разбираем в блоке про CIDR ---
variable "vpc_cidr" {
  description = "Подсеть для виртуальных машин (нод кластера)"
  type        = string
  default     = "10.10.0.0/24" # 256 адресов, ~251 доступно (YC резервирует часть)
}

# ⚠️ Pod CIDR и Service CIDR НЕ задаются в Terraform — это параметры Kubespray.
# Держим их здесь только как справку, чтобы видеть все четыре диапазона рядом
# и проверить, что они не пересекаются с vpc_cidr:
#   Pod CIDR      10.233.64.0/18   (дефолт Kubespray)
#   Service CIDR  10.233.0.0/18    (дефолт Kubespray)
# 10.10.0.0/24 и 10.233.x.x не пересекаются — всё в порядке.

variable "control_plane_resources" {
  description = "Ресурсы для control-plane нод (node1/node2)"
  type        = object({ cores = number, memory = number, disk = number })
  default     = { cores = 4, memory = 8, disk = 40 }
  # Требования Kubespray к control-plane: минимум 2 CPU / 4 ГБ. Берём с запасом:
  # etcd на той же ноде + системные компоненты.
}

variable "worker_resources" {
  description = "Ресурсы для воркеров"
  type        = object({ cores = number, memory = number, disk = number })
  default     = { cores = 2, memory = 4, disk = 30 }
}

variable "load_balancer_resources" {
  description = "Ресурсы для внешних HAProxy/keepalived нод (lb1/lb2)"
  type        = object({ cores = number, memory = number, disk = number })
  default     = { cores = 2, memory = 2, disk = 20 }
}

variable "ssh_public_key_path" {
  description = "Путь к публичному ssh-ключу"
  type        = string
  default     = "~/.ssh/id_ed25519.pub"
}

variable "ssh_user" {
  description = "Пользователь на ВМ (под ним ходит Ansible)"
  type        = string
  default     = "ubuntu"
}

variable "allowed_ssh_cidr" {
  description = "Единственный доверенный CIDR для SSH, обычно текущий публичный IP /32"
  type        = string

  validation {
    condition     = can(cidrhost(var.allowed_ssh_cidr, 0)) && var.allowed_ssh_cidr != "0.0.0.0/0" && var.allowed_ssh_cidr != "::/0"
    error_message = "Укажи ограниченный CIDR, например 203.0.113.42/32; 0.0.0.0/0 и ::/0 запрещены."
  }
}
