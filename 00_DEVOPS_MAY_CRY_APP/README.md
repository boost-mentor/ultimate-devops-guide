# DevOps May Cry

Сквозное приложение видео №2: диспетчерская ночных дежурных принимает заказы на
устранение «демонов»-инцидентов. Визуальный референс — готический neon-action,
но вселенная, названия и тексты полностью оригинальные.

Это не код внутри Kubernetes YAML. Здесь отдельно живут Go API, тесты,
multi-stage Dockerfile, PostgreSQL-миграция, Compose и Kubernetes-манифесты.

## Локальная проверка

```bash
cp .env.example .env
# Замени учебный пароль в .env. Файл .env не коммитится.
go test ./...
go test -race ./...
go vet ./...
docker compose up --build -d
docker compose ps
docker compose logs --tail=20 api
curl http://127.0.0.1:8080/livez
curl http://127.0.0.1:8080/readyz
curl http://127.0.0.1:8080/quote
curl -H 'Content-Type: application/json' \
  -d '{"target":"dns-ghost"}' \
  http://127.0.0.1:8080/orders
docker compose exec postgres psql -U devopsmaycry -d devopsmaycry \
  -c 'SELECT id, target, status, handled_by FROM orders ORDER BY id DESC LIMIT 3;'
docker compose down -v
```

Здесь используются обычные, читаемые команды без `jq`, регулярных выражений и
скрытых скриптов. `--tail=20` ограничивает журнал последними двадцатью строками.
Заголовок `Content-Type` сообщает серверу, что тело запроса — JSON; `-d`
передаёт это тело и поэтому `curl` выполняет `POST`. `dns-ghost` — вымышленное
название инцидента, а не особый режим DNS.

Последний `SELECT` — доказательство сохранения заявки в PostgreSQL. Named volume
`postgres-data` переживает обычный `docker compose down`; ключ `-v` в финальной
команде удаляет volume намеренно и используется только при полном сбросе лабы.

Публикуемый образ собирается для `linux/amd64` и `linux/arm64`, получает OCI
metadata (`version`, commit, build time), SBOM/provenance и в Kubernetes
фиксируется по digest. `latest` в учебном и production-маршруте не используется.

## API, которое используется в лабораторных

- `GET /` — HTML в браузере либо JSON-фраза для простого клиента;
- `GET /quote` — JSON-фраза и имя экземпляра, который ответил;
- `GET /status` — имя экземпляра и активный тип хранилища;
- `GET /order?on=dns-ghost` — совместимый учебный GET-вариант;
- `POST /orders` — основной контракт создания заявки; успешный код — `201`;
- `GET /overload?sec=60` — управляемая CPU-нагрузка для изолированной лабы;
- `GET /wound?mb=200` — удержание памяти для OOM-лабы;
- `GET /healthz` и `GET /livez` — жив ли HTTP-процесс;
- `GET /readyz` — может ли экземпляр обслужить запрос прямо сейчас;
- `GET /metrics` — счётчики запросов и принятых заявок.

`/livez` и `/readyz` — не магические имена Kubernetes, а HTTP endpoints самого
приложения. Kubernetes обращается к ним через `livenessProbe` и
`readinessProbe`. Ошибка liveness приводит к перезапуску контейнера; ошибка
readiness временно убирает Pod из Service endpoints. Поэтому readiness включает
проверку активного хранилища, а liveness не зависит от PostgreSQL.

Без `DB_HOST` API использует память: это удобно в лабораториях по HPA и сети.
Compose подключает PostgreSQL. Локальный DB-вариант в Kubernetes нужен для
демонстрации scheduling; в production базу выносят в managed service или
оператор, используют CSI-диски, backup/PITR и отдельный failure domain.

Самонагрузочные `/overload` и `/wound` по умолчанию выключены. Они включаются
только явным `LAB_ENDPOINTS_ENABLED=true` в изолированном стенде; публично
оставлять такие ручки нельзя.

В Part 1 Kubernetes-манифесты находятся в `k8s/base/`: Deployment запускает
несколько экземпляров API, Service даёт им постоянное имя, а probes используют
`/livez` и `/readyz`. PostgreSQL в кластер пока не переносится: API работает с
памятью, потому что цель этой части — собрать кластер и обеспечить доступ к
приложению. Compose с PostgreSQL остаётся локальной контрольной точкой.

В следующих частях учебный PostgreSQL добавляется отдельными манифестами. Это не
готовая PostgreSQL HA: в production предпочтительны managed database или
оператор, TLS, backup/PITR, CSI storage и отдельный failure domain.
