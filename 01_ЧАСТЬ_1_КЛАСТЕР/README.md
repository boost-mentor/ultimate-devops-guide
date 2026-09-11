# Блок 2 · Часть 1 — откуда берётся кластер

Практика к рабочей тетради **«Production-кластер — лабораторная тетрадь» (Блок 2)** и к видео
«Откуда берётся production-кластер». Папки открываются по номеру — в том же порядке, что в видео и тетради.

| Папка | Что внутри | В тетради |
|---|---|---|
| `01_1.1_hello_terraform` | Terraform без облака: цикл init → plan → apply на локальных файлах, state, drift | раздел 03 |
| `02_1.2_hello_ansible` | Маленький Ansible-проект: inventory, роль `dispatcher`, идемпотентность, precedence переменных, handler | раздел 04 |
| `03_1.2bis_kubespray_anatomy` | Как устроен Kubespray (tag v2.31.0), какие файлы за что отвечают, ссылки на первоисточники | раздел 05 |
| `04_1.3_five_vm_terraform` | Terraform стенда: сеть, security group с allowlist `/32`, пять VM, `output inventory` → `inventory.ini` | раздел 06 |
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

```bash
# 1. Локальные лаборатории — без облака
cd 01_1.1_hello_terraform   # terraform init && terraform plan -var-file=env/demo.tfvars && terraform apply
cd ../02_1.2_hello_ansible  # ansible-playbook playbooks/site.yml (дважды: changed → changed=0)

# 2. Стенд в облаке
export YC_TOKEN=$(yc iam create-token)
export YC_CLOUD_ID=$(yc config get cloud-id)
export YC_FOLDER_ID=$(yc config get folder-id)
cd ../04_1.3_five_vm_terraform/terraform
cp env/video.tfvars.example env/video.tfvars     # свой SSH-ключ и свой IP /32
terraform init && terraform plan -var-file=env/video.tfvars -out=video2.tfplan
terraform apply video2.tfplan
terraform output -raw inventory > ../../07_1.5_kubespray_full/inventory/video2/inventory.ini

# 3. Preflight → Kubespray
cd ../../06_1.4bis_node_preflight && ansible-playbook playbooks/preflight.yml
git clone --branch v2.31.0 --depth 1 https://github.com/kubernetes-sigs/kubespray.git
# скопировать inventory/video2 из 07_1.5_kubespray_full в клон и запустить cluster.yml — см. README папок 03 и 07

# 4. Managed-кластер рядом
cd ../05_1.4_managed_cluster/terraform
cp private.auto.tfvars.example private.auto.tfvars   # свой IP /32, свой SSH-ключ
terraform init && terraform apply
```

## Личные значения не коммитятся

`*.tfvars` (кроме `*.example`), `.env`, сгенерированный `inventory.ini`, `admin.conf`/kubeconfig, `*.tfplan`, `.terraform/` —
в `.gitignore`. В примерах стоят документационные адреса `203.0.113.x`; подставь свои.

## Снести стенд

```bash
terraform destroy   # в 04_1.3_five_vm_terraform/terraform и в 05_1.4_managed_cluster/terraform
```
