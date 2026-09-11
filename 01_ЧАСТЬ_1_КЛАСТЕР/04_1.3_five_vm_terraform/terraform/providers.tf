# Provider получает учётные данные из окружения, а не из HCL и state.
provider "yandex" {
  zone = var.zone
}
