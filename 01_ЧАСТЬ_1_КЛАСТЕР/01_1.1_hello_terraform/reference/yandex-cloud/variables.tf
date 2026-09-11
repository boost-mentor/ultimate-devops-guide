variable "zone" {
  description = "Зона доступности для подсети и VM"
  type        = string
  default     = "ru-central1-a"
}

variable "subnet_id" {
  description = "Идентификатор уже созданной подсети"
  type        = string
}

variable "image_id" {
  description = "Идентификатор образа Linux"
  type        = string
}

