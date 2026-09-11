locals {
  # Local — внутреннее вычисление, а не значение пользователя. Все артефакты
  # складываем в build рядом с конфигурацией.
  output_dir = "${path.module}/build"

  # Собираем один объект для JSON-шаблона, чтобы не повторять выражения.
  routing_summary = {
    office      = var.office_name
    dispatcher  = var.dispatcher
    shift       = var.shift
    route_count = length(var.routes)
    routes      = var.routes
  }
}
