# Адреса берём из созданных VM, состав групп — из ролей в local.nodes.
# В этом файле нет отдельных списков имён машин, которые надо править вручную.
output "kubespray_inventory" {
  value = <<-EOT
[all]
%{for name, node in module.linux_vm.nodes~}
%{if !contains(node.roles, "load-balancer")~}
${name} ansible_host=${node.external_ip} ip=${node.internal_ip} access_ip=${node.internal_ip} ansible_user=${var.ssh_user}
%{endif~}
%{endfor~}

[kube_control_plane]
%{for name, node in module.linux_vm.nodes~}
%{if contains(node.roles, "control-plane")~}
${name}
%{endif~}
%{endfor~}

[etcd]
%{for name, node in module.linux_vm.nodes~}
%{if contains(node.roles, "etcd")~}
${name}
%{endif~}
%{endfor~}

[kube_node]
%{for name, node in module.linux_vm.nodes~}
%{if contains(node.roles, "worker")~}
${name}
%{endif~}
%{endfor~}

[k8s_cluster:children]
kube_control_plane
kube_node
  EOT
}

output "kubespray_all_yml" {
  value = <<-EOT
---
supplementary_addresses_in_ssl_keys:
%{for name in local.api_endpoint_nodes~}
  - ${module.linux_vm.nodes[name].external_ip}
%{endfor~}
  EOT

  precondition {
    condition = length(local.api_endpoint_nodes) == 1 && alltrue([
      for name in local.api_endpoint_nodes : contains(local.nodes[name].roles, "control-plane")
    ])
    error_message = "В local.nodes укажи api_endpoint = true ровно у одной машины с ролью control-plane."
  }
}

output "external_ha_inventory" {
  value = <<-EOT
[load_balancers]
%{for name, node in module.linux_vm.nodes~}
%{if contains(node.roles, "load-balancer")~}
${name} ansible_host=${node.external_ip} private_ip=${node.internal_ip} ha_overlay_ip=${local.nodes[name].ha.overlay_ip} ha_priority=${local.nodes[name].ha.priority}
%{endif~}
%{endfor~}

[kubernetes_nodes]
%{for name, node in module.linux_vm.nodes~}
%{if !contains(node.roles, "load-balancer")~}
${name} ansible_host=${node.external_ip} private_ip=${node.internal_ip}
%{endif~}
%{endfor~}

[all:vars]
ansible_user=${var.ssh_user}
ansible_python_interpreter=/usr/bin/python3
ha_vip=10.77.0.10
ha_vip_prefix=24
ha_frontend_port=8080
k8s_nodeport=30080
  EOT
}
