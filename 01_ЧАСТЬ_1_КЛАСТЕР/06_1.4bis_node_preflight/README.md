# Preflight нод перед Kubespray

Эта лабораторная отвечает только на вопрос «можно ли безопасно запускать
Kubespray на этих Linux-машинах сейчас?». Она не устанавливает пакеты, не
переключает sysctl и не чинит сеть автоматически.

Открывать файлы по порядку:

1. `inventory/hosts.ini` — безопасные placeholders для трёх нод;
2. `playbooks/preflight.yml` — play для группы `k8s_cluster`;
3. `roles/node_preflight/tasks/main.yml` — read-only проверки ОС, Python, swap,
   kernel modules и IPv4 forwarding.

Реальный inventory, созданный Terraform, содержит стандартные Kubespray-группы
`kube_control_plane`, `etcd`, `kube_node` и агрегирующую группу `k8s_cluster`.
Именно к последней применяется preflight, поэтому проверяются все ноды, которые
получат компоненты Kubernetes.

Красный assert — стоп-сигнал, а не повод повторять playbook. Сначала читают
конкретное условие и выясняют владельца изменения. Например, network route и
allowlist могут принадлежать платформенной команде, а sysctl — владельцу образа.
Recovery доказывается повторным зелёным preflight и только после этого запуском
Kubespray; одного изменения конфигурации недостаточно.
