terraform {
  # Ограничиваем версию CLI: лаборатория не зависит от случайного будущего
  # major-релиза Terraform.
  required_version = ">= 1.9.0, < 2.0.0"

  required_providers {
    local = {
      # Provider local умеет управлять файлами на этом компьютере. На нём
      # безопасно показываем тот же цикл, который позже применим к Yandex Cloud.
      source  = "hashicorp/local"
      version = "2.5.2"
    }
  }
}
