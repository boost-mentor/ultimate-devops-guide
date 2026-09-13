<div align="center">

# Ultimate DevOps Guide

**Практика к серии видео-гайдов: от `docker run` до production-кластера, поднятого своими руками.**

Клонируешь, повторяешь за видео, ломаешь и чинишь. Ничего не «уже сделано за кадром».

[![Видео](https://img.shields.io/badge/YouTube-гайды-FF0000?logo=youtube&logoColor=white)](https://youtube.com/@Viktor.Golang)
[![Тетрадь](https://img.shields.io/badge/Рабочая_тетрадь-PDF-0c8aad)](https://boostmentor.ru/guides/workbook2-part1)
[![Telegram](https://img.shields.io/badge/Telegram-канал-229ED9?logo=telegram&logoColor=white)](https://t.me/booostmentor)
[![Платформа](https://img.shields.io/badge/BoostMentor-обучение-5a49e6)](https://boostmentor.ru)

</div>

---

## Два блока

| Блок | Папки | О чём | Видео |
|---|---|---|---|
| **1 · Kubernetes** | [`app/`](app) · [`lab/`](lab) | Docker, сети, кластер через Kubespray на 3 VM, все сущности, `kubectl apply` по шагам | [гайд 5 ч](https://youtu.be/6mhlHDJOQAw) |
| **2 · Production-кластер** | [`00_DEVOPS_MAY_CRY_APP/`](00_DEVOPS_MAY_CRY_APP) · [`01_ЧАСТЬ_1_КЛАСТЕР/`](01_ЧАСТЬ_1_КЛАСТЕР) | Terraform → Ansible → Kubespray: пять VM в облаке, свой кластер и managed рядом, одно приложение в оба | часть 1 вышла |

К каждому блоку есть рабочая тетрадь: команды с ожидаемым выводом, разбор ошибок и вопросы с собеседований. Забирается по ссылке в описании видео.

---

## Блок 2 · Откуда берётся production-кластер

Полный путь от пустого облака до кластера, в который заезжает приложение.

```
00_DEVOPS_MAY_CRY_APP/     сквозное приложение: Go + PostgreSQL, Compose, multistage, манифесты
01_ЧАСТЬ_1_КЛАСТЕР/
├── 01_1.1_hello_terraform      цикл init → plan → apply, state, drift — без облака и без счёта
├── 02_1.2_hello_ansible        inventory, роль, идемпотентность, precedence переменных
├── 03_1.2bis_kubespray_anatomy как устроен Kubespray, tag v2.31.0
├── 04_1.3_five_vm_terraform    стенд: сеть, security group /32, пять VM, outputs → inventory
├── 05_1.4_managed_cluster      managed-кластер тем же языком: сервисные аккаунты, node group
├── 06_1.4bis_node_preflight    проверка Linux до установки + подготовка модулей и sysctl
├── 07_1.5_kubespray_full       inventory и group_vars: containerd, Calico, kube-proxy, etcd
├── 08_1.6_verify_cluster       манифесты приложения: PSA, probes, topology spread
├── 09_1.7_cidr_to_node         Calico: доказательство связи между нодами
├── 12_1.10_compare_clusters    те же манифесты как base + overlays под оба кластера
└── 10 · 11 · 13                следующая часть: upgrade, MetalLB, внешний HA-вход
```

Порядок прохождения, команды и что понадобится — в [`01_ЧАСТЬ_1_КЛАСТЕР/README.md`](01_ЧАСТЬ_1_КЛАСТЕР/README.md).

> **Стенд платный.** Пять VM в Yandex Cloud стоят около 30 ₽ в час. Новым пользователям дают стартовый грант, его хватает примерно на 135 часов. Прошёл раздел — `terraform destroy`, и счёт останавливается.

---

## Блок 1 · Kubernetes с нуля

### [`app/`](app) — приложение для практики
Сервис на Go отдаёт цитаты и ходит в Postgres по имени сервиса. На нём разбираем Docker и сети:

- `Dockerfile.simple` — наивная сборка, образ ~302 MB;
- `Dockerfile` — multistage + Alpine, образ ~18 MB, в 17 раз меньше;
- `net-demo/docker-compose.yml` — app, postgres и netshoot в одной сети: DNS по имени, маршруты, veth вживую;
- `k8s/` — манифесты, чтобы развернуть это же приложение в кластере.

### [`lab/`](lab) — сущности по папкам
Каждая папка это одна тема, внутри README и готовые манифесты.

| Папка | Тема |
|---|---|
| `00-namespace` … `14-sidecar` | Namespace · Pod · ReplicaSet · Deployment · Service · Ingress · ConfigMap/Secret · StatefulSet · DaemonSet · Job/CronJob · Probes · Resources · RBAC · Storage · Sidecar |
| `99-final-exercise` | финальное задание, собрать всё вместе |
| `cluster-setup` | скрипты подготовки нод |

---

## Быстрый старт

**Docker и сети:**

```bash
cd app
docker build -f Dockerfile.simple -t scott:simple .    # ~302 MB
docker build -t scott:1.0.0 .                          # ~18 MB
docker images | grep scott                             # сравни размеры

cd net-demo && docker compose up -d
docker compose exec netshoot bash                      # getent hosts postgres, ip route get ...
```

**Своя инфраструктура (блок 2):**

```bash
export YC_TOKEN="$(yc iam create-token)"
export YC_CLOUD_ID="$(yc config get cloud-id)"
export YC_FOLDER_ID="$(yc config get folder-id)"

cd 01_ЧАСТЬ_1_КЛАСТЕР/01_1.1_hello_terraform          # сначала без облака
terraform init && terraform apply -var-file=env/demo.tfvars
```

## Что понадобится

Docker и `kubectl` для блока 1. Для блока 2 ещё Terraform, Ansible, `yc` и аккаунт в облаке с правом создавать ресурсы.

## Пароли здесь учебные

`DB_PASSWORD: superpass`, `API_TOKEN: demo-token` и подобное это демо-значения. На них в гайде показываем, что `base64` в Secret это кодировка, а не шифрование. В проде так не носят: encryption-at-rest для etcd, RBAC на чтение секретов, Vault или External Secrets.

Личные значения (`*.tfvars`, `.env`, kubeconfig, сгенерированный `inventory.ini`) в репозиторий не попадают, в примерах стоят документационные адреса `203.0.113.x`.

---

<div align="center">

**Виктор Шутов** · DevOps-инженер и ментор

[YouTube](https://youtube.com/@Viktor.Golang) · [Telegram](https://t.me/booostmentor) · [boostmentor.ru](https://boostmentor.ru)

</div>
