# Kubespray v2.31.0: официальный путь обновления

Источник: официальный репозиторий `kubernetes-sigs/kubespray`, tag `v2.31.0`,
commit `1c9add48975060f45396b34d8e022c30d7f80dab`.

Документация `docs/operations/upgrades.md` направляет обычное обновление через
корневой `upgrade-cluster.yml`. Этот файл импортирует
`playbooks/upgrade_cluster.yml`; собственный самодельный upgrade-скрипт в
лабораторной не используется.

Для этой сцены:

1. фиксируем зелёное состояние до обновления;
2. создаём и читаем metadata snapshot etcd;
3. меняем только `kube_version: 1.34.7` на `1.35.4`;
4. запускаем официальный `upgrade-cluster.yml`;
5. повторяем тот же набор проверок после обновления.

Ссылка: <https://github.com/kubernetes-sigs/kubespray/blob/v2.31.0/docs/operations/upgrades.md>
