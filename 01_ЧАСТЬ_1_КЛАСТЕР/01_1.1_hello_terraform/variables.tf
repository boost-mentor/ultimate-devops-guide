# Variables — входы root module. Значения лежат отдельно в env/demo.tfvars,
# поэтому один код можно запускать для другой смены без правки resources.tf.
variable "office_name" {
  description = "Название диспетчерской"
  type        = string
}

variable "dispatcher" {
  description = "Имя дежурного, которое попадёт в dispatcher.env"
  type        = string

  validation {
    # Ошибку входных данных останавливаем до создания файлов.
    condition     = length(trimspace(var.dispatcher)) >= 2
    error_message = "Имя дежурного должно содержать минимум два символа."
  }
}

variable "shift" {
  description = "Время учебной смены"
  type        = string
}

variable "routes" {
  description = "Каталог инцидентов и их стоимости"
  type        = map(number)

  validation {
    # В карте должны быть минимум два инцидента с положительной стоимостью.
    condition     = length(var.routes) >= 2 && alltrue([for price in values(var.routes) : price > 0])
    error_message = "Нужны хотя бы два маршрута с положительной стоимостью."
  }
}
