# Пины версий. Без них завтрашний apply может дать другой результат.
# Конфигурация проверена с Terraform 1.x и Yandex provider 0.219.0.
# required_version намеренно диапазоном. Провайдер запинен ЖЁСТКО: без этого
# следующий init может принести другую версию и другой plan.
terraform {
  # use_lockfile в production backend example требует Terraform >= 1.10.
  required_version = ">= 1.10.0, < 2.0.0"

  required_providers {
    yandex = {
      source  = "yandex-cloud/yandex"
      version = "0.219.0"
    }
  }
}
