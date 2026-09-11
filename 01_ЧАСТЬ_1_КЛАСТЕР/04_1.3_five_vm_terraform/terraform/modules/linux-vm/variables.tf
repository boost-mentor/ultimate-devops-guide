variable "cluster_name" {
  type = string
}

variable "nodes" {
  type = map(object({
    roles = set(string)
    resources = object({
      cores  = number
      memory = number
      disk   = number
    })
  }))
}

variable "zone" { type = string }
variable "image_id" { type = string }
variable "subnet_id" { type = string }
variable "security_group_ids" { type = list(string) }
variable "ssh_user" { type = string }
variable "ssh_public_key" {
  type      = string
  sensitive = true
}
