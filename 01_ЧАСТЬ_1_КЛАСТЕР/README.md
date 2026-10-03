# Блок 2 · исходники VIDEO2A и VIDEO2B

Практика к рабочей тетради **«Production-кластер — лабораторная тетрадь» (Блок 2)** и к видео
«Откуда берётся production-кластер». Каталоги сохранены под именами из записи.
Маршрут второй части и её отдельный релиз — в [VIDEO2B.md](../VIDEO2B.md).

| Папка | Что внутри | В тетради |
|---|---|---|
| `01_1.1_hello_terraform` | Terraform без облака: цикл init → plan → apply на локальных файлах, state, drift | раздел 03 |
| `02_1.2_hello_ansible` | Маленький Ansible-проект: inventory, роль `dispatcher`, идемпотентность, precedence переменных, handler | раздел 04 |
| `03_1.2bis_kubespray_anatomy` | Как устроен Kubespray (tag v2.31.0), какие файлы за что отвечают, ссылки на первоисточники | раздел 05 |
| `04_1.3_five_vm_terraform` | Terraform стенда: сеть, security group с allowlist `/32`, пять VM, `output kubespray_inventory` → `inventory.ini` | раздел 06 |
| `05_1.4_managed_cluster` | Managed-кластер Yandex Cloud через Terraform: сервисные аккаунты, cluster, node group | раздел 07 |
| `06_1.4bis_node_preflight` | Preflight нод перед Kubespray (проверка Linux) + `prepare-network.yml` для модулей/sysctl | раздел 08 |
| `07_1.5_kubespray_full` | Inventory и `group_vars` для Kubespray: containerd, Calico, kube-proxy iptables, etcd host | раздел 08 |
| `08_1.6_verify_cluster` | Манифесты приложения для проверки кластера: Namespace с PSA-метками, Deployment, Service, client pod | раздел 10 |
| `09_1.7_cidr_to_node` | Calico: доказательство связи между нодами | раздел 10 |
| `12_1.10_compare_clusters` | Те же манифесты как kustomize base + overlays для self-managed и managed | разделы 10–11 |
| `10_1.8_safe_upgrade` · `11_1.9_metallb` · `13_1.11_external_ha` | Следующая часть серии: безопасное обновление, MetalLB, внешний HA-вход (HAProxy + keepalived) | Блок 3 |

## Что понадобится

- **Terraform** и **Yandex Cloud CLI (`yc`)** с аккаунтом, где можно создавать ресурсы (стенд платный: пять VM + managed-кластер);
- **Ansible** (ставится в `.venv` Kubespray) и **kubectl**;
- SSH-ключ для доступа к VM.

## Порядок

1. Первая часть: локальные лаборатории Terraform и Ansible, знакомство с
   Kubespray, затем [создание пяти VM и выгрузка inventory](04_1.3_five_vm_terraform/terraform/README.md).
2. Вторая часть: managed-кластер для сравнения, preflight своих VM, установка
   Kubespray, приложение и сравнение кластеров. Открывай [карту файлов VIDEO2B](../VIDEO2B.md).

Официальный Kubespray клонируется один раз в `03_1.2bis_kubespray_anatomy/kubespray/`.
`07_1.5_kubespray_full/` хранит наши параметры и inventory. Существующие личные
параметры не заменяй образцами при повторном проходе. Команды из видео и тетради
выполняй из указанного там каталога: относительный путь зависит от него.

## Личные значения не коммитятся

Рабочие `env/video.tfvars`, `private.auto.tfvars`, `.env`, сгенерированный
`inventory.ini`, `admin.conf`/kubeconfig, state, `*.tfplan` и `.terraform/`
не публикуем. Общие `vars.auto.tfvars` и локальные учебные `env/demo.tfvars`
хранятся в Git намеренно. В адресных примерах стоят `203.0.113.x`; подставь свои.

## Снести стенд

Удаление выполняется после всей нужной практики, отдельно для пяти VM и managed.
Для пяти VM порядок с проверкой плана и нужным `-var-file` указан в
[README Terraform-стенда](04_1.3_five_vm_terraform/terraform/README.md).
Проверь, что удаляешь именно учебный стенд, и после удаления посмотри в облаке,
не осталось ли платных ресурсов.
