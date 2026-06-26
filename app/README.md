# Scott Pilgrim Quotes — демо для курса по Kubernetes

Крошечное Go-приложение, которое выдаёт случайную цитату в духе «Скотта Пилигрима против всех»
в формате `персонаж: цитата`. Один файл `main.go`, оптимизированный `Dockerfile`,
манифесты Kubernetes. Задача — наглядно показать путь **код → образ → под → сервис → доступ наружу**.

> Цитаты — короткая курированная подборка реплик из фильма. Одиночные реплики идут как
> `персонаж: цитата`, диалоги — без лейбла, с переносами строк. Добавляй свои в слайс
> `quotes` в `main.go` и пересобирай образ.

```
scott-pilgrim-k8s/
├── main.go            # всё приложение: API + фронт + опц. работа с Postgres
├── go.mod             # + github.com/lib/pq (драйвер Postgres) для модуля с БД
├── Dockerfile         # multi-stage, статический бинарь, distroless, non-root
├── .dockerignore
├── db/                # доп. модуль «приложение + БД»
│   ├── docker-compose.yml   # Postgres на worker-ноде, порт 5432, авто-миграции
│   └── init/01_init.sql     # таблица quotes + цитаты (выполняется при первом старте)
├── k8s/
│   ├── deployment.yaml          # В1: 1 под + пробы + лимиты (встроенные цитаты, без БД)
│   ├── deployment-db.yaml       # В2: то же + env DB_* и DB_PASSWORD из Secret
│   ├── secret.yaml              # В2: пароль к БД (base64 — НЕ шифрование, см. видео)
│   ├── service-clusterip.yaml   # Service ClusterIP (внутри кластера)
│   ├── service-nodeport.yaml    # Service NodePort (порт 30080 на каждой ноде)
│   ├── service.yaml             # Service LoadBalancer (EXTERNAL-IP <pending> без MetalLB)
│   └── service-externalname.yaml# Service ExternalName (DNS-CNAME наружу)
└── README.md
```

## Что внутри приложения

- `GET /` — одностраничник (HTML встроен в бинарь), фронт делает `fetch('/quote')`.
- `GET /quote` — возвращает JSON: `{"character":"рамона","quote":"...","pod":"<имя пода>"}`.
- `GET /healthz` — базовая health-ручка (жив ли процесс, отвечает ли HTTP).
- `GET /livez` — liveness probe (не ответила => k8s перезапускает под).
- `GET /readyz` — readiness probe (не готов => под выводится из балансировки Service).

Поле `pod` — это hostname контейнера. В Kubernetes это имя пода, поэтому видно,
**какой именно под** обслужил запрос (особенно наглядно, если поднять `replicas: 3`).

---

## Шаг 1. Запустить локально (без Docker)

```bash
cd scott-pilgrim-k8s
go run .
# открой http://localhost:8080
```

Что объяснить на камеру: это просто HTTP-сервер на Go. Жмём кнопку — фронт дёргает `/quote`,
бэкенд отдаёт случайную цитату. Никакого кубера тут ещё нет, обычный процесс на ноутбуке.

---

## Шаг 2. Собрать образ — два Dockerfile для сравнения

В проекте лежат ДВА Dockerfile — это самый наглядный момент про оптимизацию.

### Вариант А — простой («наивный»), `Dockerfile.simple`

Одна стадия: собираем и запускаем прямо в образе `golang`. Так пишут в начале.
Работает, но внутрь попадает весь Go-тулчейн — образ огромный.

```bash
# не дефолтное имя файла => указываем через -f
docker build -f Dockerfile.simple -t scott-pilgrim-quotes:simple .
```

### Вариант Б — оптимизированный, `Dockerfile` (дефолтный)

Multi-stage + статический бинарь + distroless + non-root.

```bash
docker build -t scott-pilgrim-quotes:1.0.0 .
```

### Сравнить размеры на камеру

```bash
docker images scott-pilgrim-quotes
```

```
REPOSITORY             TAG      SIZE
scott-pilgrim-quotes   simple   302MB     <- наивный
scott-pilgrim-quotes   1.0.0    7.53MB    <- оптимизированный (~в 40 раз меньше!)
```

Главная мысль: **код один и тот же, отличается только Dockerfile**. Оптимизация образа —
это не про приложение, а про то, как мы его упаковали.

Что показать и объяснить по оптимизированному `Dockerfile`:

1. **Multi-stage build.** Стадия `builder` (образ `golang`) содержит весь Go-тулчейн —
   компилятор, кэш модулей и т.д. В финальный образ она **не попадает**: мы копируем
   из неё только готовый бинарь. Поэтому в итоговом образе нет ни компилятора, ни исходников.
2. **Кэш зависимостей.** Сначала копируем `go.mod` и делаем `go mod download` отдельным слоём.
   Пока зависимости не меняются — Docker берёт этот слой из кэша, пересборка быстрее.
3. **Статический бинарь** — `CGO_ENABLED=0`. Бинарь ни от чего не зависит, его можно положить
   в практически пустой образ.
4. **Флаги размера** — `-ldflags="-s -w"` (убрать отладочные символы) и `-trimpath`.
5. **Минимальный рантайм** — `distroless/static`: нет shell, нет пакетного менеджера, ничего
   лишнего. Меньше размер и меньше поверхность атаки. Тег `:nonroot` => контейнер стартует
   под непривилегированным пользователем (хорошая практика безопасности).

Посмотреть, что получилось:

```bash
docker images scott-pilgrim-quotes        # размер образа — обычно единицы МБ
docker history scott-pilgrim-quotes:1.0.0  # видно слои
```

Запустить образ локально и убедиться, что внутри контейнера всё работает так же:

```bash
docker run --rm -p 8080:8080 scott-pilgrim-quotes:1.0.0
# открой http://localhost:8080
```

Что объяснить: «локально из образа» = тот же бинарь, но уже изолированный в контейнере.
Это ровно то, что поедет в Kubernetes — мы не меняем код, мы упаковали его один раз.

---

## Шаг 3. Запустить в Kubernetes — ВЕРСИЯ 1 (без БД)

> У приложения ДВЕ версии:
> **В1 — без БД** (встроенные цитаты, этот шаг) и **В2 — с БД + Secret** (Шаг 4, для демо секретов).

### 3.1 Сделать образ доступным кластеру
На РЕАЛЬНОМ кластере (3 ВМ) образ собран локально — на нодах его нет. Два пути:
```bash
# A. через registry (как в проде):
docker tag  scott-pilgrim-quotes:1.0.0  <REGISTRY>/scott-pilgrim-quotes:1.0.0
docker push <REGISTRY>/scott-pilgrim-quotes:1.0.0      # и впиши этот image в k8s/deployment.yaml
# B. импортнуть в containerd на КАЖДЫЙ воркер (без registry):
docker save scott-pilgrim-quotes:1.0.0 -o scott.tar
scp scott.tar root@<NODE_IP>:/root/ && ssh root@<NODE_IP> 'podman load -i /root/scott.tar'   # cri-o берёт образ из общего стора podman (ctr — это containerd, не наш рантайм)
```
> Локально (только упоминаем): `minikube image load …` / `kind load docker-image …`.

### 3.2 Применить манифесты (по одному — пути явные)
```bash
kubectl apply -f scott-pilgrim-k8s/k8s/deployment.yaml        # Deployment (1 под, без БД)
kubectl apply -f scott-pilgrim-k8s/k8s/service-clusterip.yaml # ClusterIP (доступ внутри кластера)
kubectl get pods -l app=scott-pilgrim -w                      # ждём Running / Ready
```
> ⚠️ НЕ `kubectl apply -f k8s/` целиком — там лежат и `deployment-db.yaml` (нужен `DB_HOST`), и другие типы сервисов.

### 3.3 Типы сервиса — отдельными манифестами (показать и применить)
```bash
kubectl apply -f scott-pilgrim-k8s/k8s/service-clusterip.yaml     # ClusterIP — только ВНУТРИ
kubectl apply -f scott-pilgrim-k8s/k8s/service-nodeport.yaml      # NodePort — порт 30080 на КАЖДОЙ ноде
kubectl apply -f scott-pilgrim-k8s/k8s/service.yaml               # LoadBalancer — EXTERNAL-IP <pending> (нет MetalLB)
kubectl apply -f scott-pilgrim-k8s/k8s/service-externalname.yaml  # ExternalName — DNS-CNAME наружу
kubectl get svc
```

### 3.4 Открыть и проверить
```bash
# универсально (везде): port-forward на сам Deployment
kubectl port-forward deploy/scott-pilgrim 8080:8080
curl localhost:8080/quote
# снаружи через NodePort (на реальном кластере) — на ЛЮБОМ воркере:
curl http://<WORKER_IP>:30080/quote
```
Что объяснить на камеру: под напрямую снаружи недоступен (внутренняя сеть). **Service** даёт стабильную
точку входа; NodePort открывает порт на всех нодах, LoadBalancer просит внешний IP у облака/MetalLB.

### 3.5 Бонус — балансировка
```bash
kubectl scale deployment scott-pilgrim --replicas=3
```
Обнови несколько раз — поле `ответил под:` меняется: Service балансирует запросы между подами.

### Убрать за собой

```bash
kubectl delete -f k8s/
```

---

## Шаг 4. ВЕРСИЯ 2 — приложение + БД + Secret (демо секретов)

Приложение умеет брать цитаты из Postgres. Включается переменной `DB_HOST` (без неё — встроенные цитаты,
все базовые демо работают как раньше). Пароль к БД — в Kubernetes Secret.

### 4.1 Поднять БД (docker compose на worker-ноде)
```bash
cd db
docker compose up -d        # Postgres + авто-миграции (init/*.sql при первом старте)
docker compose exec db psql -U scott -d scott -c 'select count(*) from quotes;'   # = 11
```

### 4.2 Secret + DB-деплой
```bash
kubectl apply -f k8s/secret.yaml
# в k8s/deployment-db.yaml впиши DB_HOST = IP воркера, где поднята БД
kubectl apply -f k8s/deployment-db.yaml
kubectl get pods -l app=scott-pilgrim -w        # Ready только когда БД доступна (/readyz пингует БД)
kubectl port-forward deploy/scott-pilgrim 8080:8080
curl localhost:8080/quote                       # цитата ИЗ Postgres
```

### 4.3 Демо «Secret — не шифрование»
```bash
kubectl get secret scott-db -o jsonpath='{.data.DB_PASSWORD}' | base64 -d; echo   # base64 != шифрование
```
Тот же пароль физически лежит в etcd открытым — серты для `etcdctl` берём из процесса
`ps -efww | grep '[e]tcd'` (kubespray: `/etc/ssl/etcd/ssl/`). Защита — encryption-at-rest на api-server.

> Полный сценарий записи модуля — `knowledge/youtube_agent/strategy/SECRETS_DB_MODULE_plan.md`.
