output "nodes" {
  value = {
    for name, instance in yandex_compute_instance.this : name => {
      external_ip = instance.network_interface[0].nat_ip_address
      internal_ip = instance.network_interface[0].ip_address
      roles       = var.nodes[name].roles
      instance_id = instance.id
    }
  }
}
