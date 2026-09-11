# Self-managed Kubernetes stand — Terraform

Один root module создаёт сеть, подсеть, security group и вызывает typed child
module `modules/linux-vm` для пяти VM: `node1`–`node3` и `lb1`–`lb2`.
Стабильные ключи `for_each` не сдвигают адреса ресурсов при изменении карты.

## Структура

- `versions.tf` — версии Terraform и провайдера;
- `providers.tf` — provider без токенов в коде;
- `variables.tf` / `terraform.tfvars.example` — входной контракт;
- `locals.tf` — карта машин, их роли, выбранная нода доступа к API и настройки HA каждой LB;
- `network.tf` / `security.tf` — сеть и доступ;
- `compute.tf` — вызов переиспользуемого модуля;
- `outputs.tf` — адреса, роли и SSH-команды созданных VM;
- `inventory.tf` — генерация inventory для Kubespray и внешнего HA из той же карты машин;
- `modules/network/` — переиспользуемые VPC и subnet;
- `modules/linux-vm/` — переиспользуемые VM, диски и сетевые интерфейсы.

Локальный state используется только в изолированной лабораторной. В production
backend описывают отдельным файлом и хранят state в versioned/encrypted storage
с locking. Такой backend здесь намеренно не изображается как уже настроенный.

## Безопасный цикл записи

```bash
cp terraform.tfvars.example terraform.tfvars
# Укажи свой публичный IP /32 и правильный SSH public key.
export YC_TOKEN="$(yc iam create-token)"
export YC_CLOUD_ID="$(yc config get cloud-id)"
export YC_FOLDER_ID="$(yc config get folder-id)"

terraform fmt -check -recursive
terraform init
terraform validate
terraform plan -out=video2.tfplan
terraform apply video2.tfplan
terraform output

# После сравнения с уже созданной учебной лабораторией:
terraform plan -destroy -out=destroy.tfplan
terraform apply destroy.tfplan
```

`terraform.tfvars`, state и plan содержат чувствительные данные и не попадают в
git. В production state должен жить в отдельном versioned/encrypted bucket с
locking и отдельным ключом на окружение; credentials приходят из CI/Vault/OIDC,
а не из `.tf`.

Адреса учебного стенда можно показать: `terraform output -raw kubespray_inventory`.
Эти outputs не содержат паролей, токенов или приватных ключей и не помечены
`sensitive`. Сам state целиком публиковать нельзя: в нём могут быть другие данные.

## От карты машин к inventory

`locals.tf` задаёт `local.nodes`. В `compute.tf` строка `nodes = local.nodes`
передаёт эту карту в модуль `linux_vm`. Внутри модуля она называется `var.nodes`;
`variables.tf` объявляет её тип, а `main.tf` использует `for_each = var.nodes`
для создания VM. `modules/linux-vm/outputs.tf` возвращает карту с адресами и
ролями. В root она доступна как `module.linux_vm.nodes`.

`inventory.tf` перебирает эту карту. Группы `kube_control_plane`, `etcd`,
`kube_node` и `load_balancers` выбирают машины по ролям `control-plane`, `etcd`,
`worker` и `load-balancer`. Имена машин в шаблоне больше не перечисляются вручную.
Чтобы добавить worker, достаточно добавить машину с ролью `worker` в `local.nodes`:
после создания VM и повторного экспорта она попадёт и в список адресов, и в группу.

У одной control-plane-машины стоит `api_endpoint = true`: её внешний IP попадёт
в дополнительный SAN сертификата. У каждой LB рядом с её ролями находятся
`ha.overlay_ip` и `ha.priority`. Эти настройки не зависят от порядка машин в карте.
VM-модуль принимает только `roles` и `resources`; дополнительные поля для
inventory остаются в root. Текущие ресурсы VM и назначение групп не изменены.

`terraform output` читает ранее сохранённые результаты, а не пересчитывает
изменённый шаблон. После изменения конфигурации сначала проверь `terraform plan`,
примени только ожидаемые изменения и повтори экспорт inventory. При добавлении
реальной машины plan будет содержать её создание — это уже изменение стенда.

Локальная регрессия без облака: `python3 tests/check_inventory.py`
(нужны Terraform, `ansible-inventory` и Python с PyYAML). Она использует временную
конфигурацию без ресурсов/провайдеров; проверяет группы через настоящий Ansible,
добавление, удаление, переименование машин, смену ролей и выбор API endpoint.

Отдельное решение про Kubernetes API описано в
`API_ENDPOINT_DECISION.md`. В Part 1 kubeconfig использует разрешённый адрес
`node1`; это учебное ограничение, а не отказоустойчивый API endpoint.
