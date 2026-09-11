# Resource — объект, жизненным циклом которого управляет Terraform. Здесь это
# три файла; в рабочем стенде такими объектами станут сеть и пять VM.
resource "local_file" "dispatcher_env" {
  filename        = "${local.output_dir}/dispatcher.env"
  file_permission = "0644"
  content = templatefile("${path.module}/templates/dispatcher.tftpl", {
    office     = var.office_name
    dispatcher = var.dispatcher
    shift      = var.shift
  })
}

resource "local_file" "journal" {
  filename        = "${local.output_dir}/journal.txt"
  file_permission = "0644"
  content = templatefile("${path.module}/templates/journal.tftpl", {
    dispatcher = var.dispatcher
    routes     = var.routes
  })
}

resource "local_file" "routing_summary" {
  filename        = "${local.output_dir}/routing-summary.json"
  file_permission = "0644"
  content = templatefile("${path.module}/templates/summary.tftpl", {
    summary_json = jsonencode(local.routing_summary)
  })

  depends_on = [
    # Сводку создаём после двух человекочитаемых файлов. Это учебный пример
    # явной зависимости; обычные ссылки на атрибуты строят её автоматически.
    local_file.dispatcher_env,
    local_file.journal,
  ]
}
