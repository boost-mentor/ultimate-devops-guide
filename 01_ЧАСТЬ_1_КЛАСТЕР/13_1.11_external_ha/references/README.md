# Учебные references — читаются, не устанавливаются

Playbook этой лаборатории не устанавливает файлы из каталога `references/`.
Это анатомические образцы: их читают в модуле reverse proxy и запускают на
локальном docker-стенде, чтобы увидеть балансировку до сборки HA-пары.

## Файлы

- `nginx-basic.conf` — минимальный NGINX reverse proxy: `events`, `http`,
  `upstream`, listener `server`, `location`, `proxy_pass`, forwarded headers.
  Работает как полный `/etc/nginx/nginx.conf` локального стенда.
- `haproxy-basic.cfg` — тот же пул словами HAProxy: `global`, `defaults`,
  `frontend`, `backend`, `balance roundrobin`, `option httpchk GET /readyz`,
  `server … check`, страница stats. Написан руками под два фиксированных
  backend'а — это его учебная роль и его же ограничение.
- `nginx-round-robin.conf` — NGINX-эквивалент лабораторного NodePort-пула
  для честного сравнения двух синтаксисов; не для apply.

## NGINX reference и граница L4/L7

`nginx-round-robin.conf` показывает тот же учебный backend pool в типичной
HTTP/L7-конфигурации NGINX. Default algorithm upstream — round robin. NGINX
завершает HTTP на себе и добавляет proxy headers; HAProxy live-конфигурация в
этой лабораторной оставляет frontend в TCP/L4 и использует HTTP только для
health check.

Reference не утверждает, что один продукт «всегда L4», а другой «всегда L7».
Оба умеют больше; выбор зависит от TLS termination, routing rules,
observability, ecosystem и операционной модели.

## Jinja: как из шаблона получается рабочий config

Статические файлы выше написаны руками. Живой конфиг лаборатории собирает
Ansible из шаблона — вот весь его путь.

- **Откуда переменные.** Умолчания — `roles/haproxy/defaults/main.yml`
  (`ha_frontend_port`, `k8s_nodeport`). Истина стенда —
  `inventory/group_vars/load_balancers.yml`, она переопределяет умолчания и
  добавляет VIP. Адреса нод шаблон берёт из `hostvars` runtime-inventory,
  который генерирует Terraform.
- **Что рендерится.** `roles/haproxy/templates/haproxy.cfg.j2`: обычный
  literal-текст плюс подстановки `{{ … }}` и цикл `{% for %}` по группе
  `kubernetes_nodes` — по строке `server` на каждую ноду.
- **Куда пишется.** В `/etc/haproxy/haproxy.cfg` на каждой LB-VM. Сначала —
  во временный файл, а не поверх живого конфига.
- **Как валидируется.** Параметр `validate: "/usr/sbin/haproxy -c -f %s"`
  запускает проверку на временном пути. Красная проверка — live-файл не
  заменяется; замена происходит атомарно и только после зелёного `-c`.
- **Какой handler.** `Reload HAProxy` (`systemd state: reloaded`). `notify`
  зовёт его только если файл реально изменился; reload перечитывает
  конфигурацию без разрыва установленных соединений — restart здесь не нужен.
- **Чем template отличается от готового config.** Файл на диске — производная.
  Истина — шаблон плюс inventory: ручная правка на VM — это дрейф, который
  следующий запуск playbook молча перепишет.
