output "nodes" {
  description = "Созданные VM: адреса и роли из карты машин"
  value       = module.linux_vm.nodes
}

output "ssh_commands" {
  description = "Ready-to-use SSH commands for access checks"
  value = [
    for _, node in module.linux_vm.nodes :
    "ssh ${var.ssh_user}@${node.external_ip}"
  ]
}

output "load_balancers" {
  description = "VM с ролью load-balancer для внешнего HA"
  value = {
    for name, node in module.linux_vm.nodes : name => node
    if contains(node.roles, "load-balancer")
  }
}
