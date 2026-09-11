# Kubespray в этом проекте

Официальный проект: <https://github.com/kubernetes-sigs/kubespray>  
Документация: <https://kubespray.io/>  
Release для лабораторной: `v2.31.0`.

Во время лабораторной официальный release клонируется в локальную папку
`kubespray/`. Эта папка намеренно не хранится в учебном репозитории: учащийся
работает с оригинальными исходниками проекта, а не с нашей копией.

```bash
git clone --branch v2.31.0 --depth 1 \
  https://github.com/kubernetes-sigs/kubespray.git kubespray
cd kubespray
git describe --tags --exact-match
```

Последняя команда должна вывести `v2.31.0`. Это фиксирует повторяемый source,
вместо меняющейся основной ветки. Далее официальный проект читается сверху вниз:

1. `README.md` и `docs/getting_started/getting-started.md` — требования и
   поддерживаемый путь установки;
2. `cluster.yml` — корневой playbook, который импортирует более узкие plays;
3. `inventory/sample/` — официальный образец inventory и group_vars;
4. `roles/kubespray_defaults/` — значения по умолчанию и проверки версий;
5. `roles/container-engine/`, `roles/etcd/`, `roles/kubernetes/` и
   `roles/network_plugin/` — крупные этапы сборки кластера.

Kubespray — не отдельная магия поверх Ansible. Это большой Ansible-проект:
inventory отвечает, какие машины получают роли control plane, etcd и worker;
group_vars отвечают, какую версию Kubernetes, runtime, CNI и CIDR установить;
`cluster.yml` связывает эти решения с готовыми roles.

## Решения именно этого стенда

В Git хранятся только решения конкретного стенда:

- `inventory/video2/inventory.example.ini` — безопасный пример топологии;
- `inventory/video2/group_vars/` — параметры Kubernetes, runtime и сети;
- `versions/` — исходная и целевая конфигурации настоящего upgrade.

Реальный `inventory.ini` создаётся из Terraform outputs после создания пяти VM.
В Kubespray входят `node1`–`node3`. Машины `lb1` и `lb2` остаются отдельным
внешним HA-слоем и настраиваются другим Ansible-проектом.

Перед запуском реальный inventory не открывается на экране: он содержит адреса.
Топология показывается на безопасном `inventory.example.ini`, затем проверяется
`ansible-inventory --graph`. Доступ, sudo и Linux prerequisites проверяются
отдельной лабораторной `06_1.4bis_node_preflight`; она ничего не исправляет.

Если SSH или маршрут недоступны, сначала фиксируются timestamp, DNS/route и
TCP/HTTPS evidence и уточняется владелец сети/allowlist. Слепое ослабление
firewall или host-key checking не считается исправлением. Временный разрешённый
маршрут допустим для записи только как workaround с обозначенным долгом.
