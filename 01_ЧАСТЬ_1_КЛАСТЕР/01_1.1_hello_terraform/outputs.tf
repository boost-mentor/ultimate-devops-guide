# Output — результат root module для человека или следующего инструмента.
# Позже output Terraform станет готовым inventory для Kubespray.
output "artifacts" {
  description = "Публичный контракт root module для следующего шага"
  value = {
    dispatcher_env  = local_file.dispatcher_env.filename
    journal         = local_file.journal.filename
    routing_summary = local_file.routing_summary.filename
  }
}
