# HAProxy + keepalived: внешний HA-слой Part 1

Первичные источники:

- HAProxy: <https://github.com/haproxy/haproxy> и
  <https://docs.haproxy.org/>;
- keepalived: <https://github.com/acassen/keepalived>;
- NGINX: <https://nginx.org/en/docs/http/load_balancing.html>.

У HAProxy и keepalived нет единственной официальной Ansible-role, которую можно
без проверки назвать production-стандартом. Поэтому лаборатория использует
короткие локальные roles поверх дистрибутивных пакетов: весь install, validate,
reload и proof виден в source. Стороннюю Galaxy-role в рабочем проекте сначала
фиксируют по версии, проверяют её defaults/tasks и прогоняют тем же acceptance,
а не принимают как чёрный ящик из-за слова «готовая».

## Порядок чтения source

1. `ARCHITECTURE_DECISION.md` — что именно доказывает внутренний VIP;
2. `inventory.example.ini` — две LB-VM и три Kubernetes-ноды без реальных IP;
3. `inventory/group_vars/load_balancers.yml` — VIP, frontend и NodePort;
4. `preflight-backends.yml` — `/readyz` каждого backend до установки HA;
5. `roles/vxlan_l2/` — persistent private L2 для VRRP;
6. `roles/haproxy/templates/haproxy.cfg.j2` — L4 frontend, round-robin
   backends, L7 health checks и loopback stats;
7. `roles/keepalived/` — VRRP, приоритеты, VIP и сильный track script;
8. `playbooks/site.yml` — единственный канонический порядок установки и proof;
9. `references/` — NGINX-конфигурация для честного сравнения, не для apply;
10. `scripts/failover-probe.sh` — открытый измеритель разрыва, не install magic.

## Почему конфигурация меняется безопасно

Ansible `template` сначала рендерит временный файл. `validate` запускает
`haproxy -c` или `keepalived -t` на временном пути. Только зелёный файл атомарно
заменяет live-конфигурацию; затем handler делает reload. Ошибка синтаксиса не
должна уничтожить работающий процесс. После reload playbook снова читает config,
проверяет status, stats и HTTP через VIP.

## L4, L7 и round robin

Frontend работает в `mode tcp`: HAProxy принимает TCP connection и выбирает
NodePort backend round-robin. TLS на таком L4 пути можно пропустить насквозь;
сертификат остаётся у приложения/Ingress. Отдельный HTTP health check обращается
к `/readyz`, поэтому открытый TCP-порт с неготовым приложением не считается
здоровым backend.

NGINX чаще встречается как web server и L7 reverse proxy, HAProxy — как
специализированный load balancer, но граница не абсолютна: HAProxy умеет HTTP/L7,
а NGINX stream module — TCP/L4. В `references/nginx-round-robin.conf` показан
эквивалентный L7 reference; он не устанавливается этой ролью.

Повторные HTTP-запросы показывают поле `pod`, но при
`externalTrafficPolicy: Cluster` один NodePort может переслать запрос на Pod
другой ноды. Поэтому доказательством HAProxy round robin служат растущие session
counters у разных backend в stats, а не только меняющееся имя Pod.

Stats привязан к `127.0.0.1:8404`. Для браузера его открывают через локальный
SSH tunnel к конкретной LB-VM; наружу без аутентификации порт не публикуется.
`playbooks/site.yml` дополнительно создаёт двенадцать новых соединений и читает
CSV stats у текущего владельца VIP. Приёмка требует положительный session
counter у каждого здорового NodePort backend — это проверка round robin на
уровне HAProxy, а не догадка по имени Pod.

Host-key checking в `ansible.cfg` включён. До первого удалённого запуска ключи
VM сверяют с доверенным источником и добавляют в `known_hosts`; `accept-new` и
отключение проверки не являются исправлением доступа.

## Измеряемый failover

До воздействия фиксируют hostname текущего владельца VIP. На одной LB запускают
`scripts/failover-probe.sh`, затем на владельце намеренно останавливают HAProxy.
Приёмка требует одновременно:

- старый владелец потерял VIP, новый владелец отличается;
- probe после измеренной серии ошибок снова получает HTTP 200;
- keepalived journal объясняет переход;
- после возврата HAProxy оба backend healthy; из-за `nopreempt` VIP может
  остаться на новом владельце — это ожидаемо.

`10.77.0.10` — внутренний адрес VXLAN. SSH tunnel делает демонстрацию доступной
с ноутбука, но не превращает VIP в production public entry. Реальный внешний
адрес требует согласованного маршрута/L2/BGP/floating IP и отдельного владельца.
