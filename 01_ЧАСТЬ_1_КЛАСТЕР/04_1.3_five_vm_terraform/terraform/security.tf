locals {
  # Вход в SSH, API и NodePort ограничен адресом оператора. Внутреннее
  # взаимодействие нод разрешает сам security group module.
  trusted_admin_cidrs = [var.allowed_ssh_cidr]
}
