variable "name_prefix" {
  type = string
}

variable "zone" {
  type = string
}

variable "vpc_cidr" {
  type = string
}

variable "trusted_admin_cidrs" {
  type = list(string)
}
