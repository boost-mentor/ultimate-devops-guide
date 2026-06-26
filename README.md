# Kubernetes Блок 1 — лаборатория · BoostMentor

Практика к гайду **«Kubernetes — рабочая тетрадь»**. Клонируешь репозиторий, поднимаешь приложение и кластер, трогаешь каждую тему руками: Docker, сети, RBAC, etcd, все сущности и `kubectl apply` по шагам.

Гайд (78 страниц, command-first) забираешь в Telegram-канале — ссылка внизу.

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

- ✈ **Telegram — [t.me/booostmentor](https://t.me/booostmentor)** — полный гайд, разборы, следующие блоки серии
- 🌐 **[boostmentor.ru](https://boostmentor.ru)** — менторство и обучение DevOps

Автор — Виктор Шутов, DevOps-инженер и ментор.
