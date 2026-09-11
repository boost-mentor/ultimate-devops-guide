# Kubernetes — лаборатория · BoostMentor

Практика к серии рабочих тетрадей **«Kubernetes с нуля, руками, до собеса»**. Клонируешь репозиторий, поднимаешь приложение и кластер, трогаешь каждую тему руками. Тетради (PDF) забираешь в Telegram-канале — ссылка внизу.

| Блок | Папки | О чём |
|---|---|---|
| **Блок 1 · Kubernetes** | `app/`, `lab/` | Docker, сети, свой кластер через kubespray на 3 VM, все сущности, `kubectl apply` по шагам |
| **Блок 2 · Production-кластер** | `00_DEVOPS_MAY_CRY_APP/`, `01_ЧАСТЬ_1_КЛАСТЕР/` | Terraform → Ansible → Kubespray: пять VM в облаке, свой кластер и managed рядом, одно приложение в оба |

---

## Блок 2 · Production-кластер — откуда берётся кластер

- `00_DEVOPS_MAY_CRY_APP/` — сквозное приложение DEVOPS MAY CRY (Go + PostgreSQL): Compose для локального запуска, multi-stage Dockerfile, манифесты Kubernetes.
- `01_ЧАСТЬ_1_КЛАСТЕР/` — тринадцать лабораторных по порядку видео: Hello Terraform → Hello Ansible → анатомия Kubespray → Terraform пяти VM → managed-кластер → preflight → Kubespray → проверка приложением → сравнение кластеров. Порядок, команды и что понадобится — в [`01_ЧАСТЬ_1_КЛАСТЕР/README.md`](01_ЧАСТЬ_1_КЛАСТЕР/README.md).

Стенд платный (пять VM + managed-кластер в Yandex Cloud) — после практики `terraform destroy`. Личные значения (`*.tfvars`, `.env`, kubeconfig, сгенерированный `inventory.ini`) не коммитятся; в примерах документационные адреса `203.0.113.x`.

---

## Блок 1 · Kubernetes — лаборатория

## Что внутри

```
app/   — демо-приложение scott-pilgrim (Go + Postgres)
lab/   — пошаговый курс по сущностям Kubernetes
```

### `app/` — приложение для практики
Сервис на Go отдаёт цитаты и ходит в Postgres по имени сервиса. На нём разбираем Docker и сети:
- `Dockerfile.simple` — наивная сборка, образ ~302 MB;
- `Dockerfile` — multistage + Alpine, образ ~18 MB (×17 меньше);
- `net-demo/docker-compose.yml` — app + postgres + netshoot в одной сети, чтобы вживую посмотреть DNS по имени, маршруты и veth;
- `k8s/` — манифесты, чтобы развернуть это же приложение в кластере.

### `lab/` — сущности по папкам
Каждая папка = одна тема, внутри README и готовые манифесты:

| Папка | Тема |
|---|---|
| `00-namespace` … `14-sidecar` | Namespace · Pod · ReplicaSet · Deployment · Service · Ingress · ConfigMap/Secret · StatefulSet · DaemonSet · Job/CronJob · Probes · Resources · RBAC · Storage · Sidecar |
| `99-final-exercise` | финальное задание — собрать всё вместе |
| `cluster-setup` | скрипты подготовки нод |

## Что понадобится

- **Docker** — для `app/` и разбора сети;
- **Kubernetes-кластер** — для `lab/`. Свой кластер на 3 VM через kubespray поднимаешь по гайду (раздел «Раскатка»); для большинства тем хватит и minikube;
- **kubectl**.

## Быстрый старт

**Приложение и сеть (Docker):**
```bash
cd app
docker build -f Dockerfile.simple -t scott:simple .   # ~302 MB
docker build -t scott:1.0.0 .                          # ~18 MB
docker images | grep scott                             # сравни размеры

cd net-demo
docker compose up -d
docker compose exec netshoot bash                      # внутри: getent hosts postgres, ip route get ...
```

**Сущности (Kubernetes):**
```bash
kubectl apply -f lab/00-namespace/
kubectl apply -f lab/01-pod/
# дальше по папкам — в каждой README объясняет, что смотреть
```

## Пароли здесь учебные

`DB_PASSWORD: superpass`, `API_TOKEN: demo-token` и подобное — **демо-значения для обучения**. На них в гайде показываем, что `base64` в Secret это кодировка, а не шифрование. В проде так не носят: encryption-at-rest для etcd, RBAC на чтение секретов, Vault / External Secrets.

## Гайд и продолжение

- ✈ **Telegram — [t.me/booostmentor](https://t.me/booostmentor)** — рабочие тетради обоих блоков, разборы, следующие блоки серии
- 🌐 **[boostmentor.ru](https://boostmentor.ru)** — менторство и обучение DevOps

Автор — Виктор Шутов, DevOps-инженер и ментор.
