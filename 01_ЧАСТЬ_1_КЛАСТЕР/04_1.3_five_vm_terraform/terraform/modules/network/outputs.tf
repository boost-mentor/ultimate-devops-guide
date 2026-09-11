output "network_id" {
  value = yandex_vpc_network.this.id
}

output "subnet_id" {
  value = yandex_vpc_subnet.this.id
}

output "security_group_id" {
  value = yandex_vpc_security_group.this.id
}
