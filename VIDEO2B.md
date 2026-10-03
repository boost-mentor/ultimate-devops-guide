# VIDEO2B · от виртуальных машин до Kubernetes и приложения

Исходники второго выпуска серии Terraform → Ansible → Kubespray.
Версия материалов: **`video2b-v1.0`**.

В [первом выпуске](https://youtu.be/Lyc7Ng66GO0) разобрали Terraform и Ansible,
создали пять VM и получили inventory из Terraform outputs. Здесь продолжаем:
создаём managed-кластер для сравнения, проверяем Linux на своих VM, устанавливаем
Kubernetes через Kubespray и запускаем приложение в обоих кластерах.

## Скачать именно эту версию

[Релиз с архивом исходников](https://github.com/boost-mentor/ultimate-devops-guide/releases/tag/video2b-v1.0).
На странице релиза выбери **Source code (zip)**. Это полный снимок общего
репозитория: предыдущие примеры оставлены, чтобы можно было восстановить стенд.

Для нового рабочего каталога можно вместо ZIP клонировать тег:

```bash
git clone --branch video2b-v1.0 --depth 1 https://github.com/boost-mentor/ultimate-devops-guide.git
cd ultimate-devops-guide
```

Если уже работаешь в своём клоне, **не заменяй его новым архивом поверх файлов**:
там могут быть твой Terraform state, значения параметров и inventory. Сравни
нужные исходники отдельно и сохрани существующее состояние стенда.

## Что за приложение

[DEVOPS MAY CRY](00_DEVOPS_MAY_CRY_APP/README.md) — учебная диспетчерская:
HTTP API на Go принимает заявки на вымышленные инциденты. В локальном Compose
оно использует PostgreSQL. В манифестах этого выпуска запускается API без
`DB_HOST`, с хранилищем в памяти; PostgreSQL в эти кластеры здесь не устанавливаем.

## Какие файлы нужны во второй части

Пути в таблице указаны относительно `01_ЧАСТЬ_1_КЛАСТЕР/`.
Старая нумерация папок сохранена, чтобы пути совпадали с записью.

| Этап | Что открыть |
|---|---|
| Managed Kubernetes | [`05_1.4_managed_cluster/terraform/`](01_ЧАСТЬ_1_КЛАСТЕР/05_1.4_managed_cluster/terraform): `main.tf`, `network.tf`, `service-accounts.tf`, `kubernetes.tf`, `outputs.tf`, публичные `vars.auto.tfvars` и образец личных параметров |
| Linux перед установкой | [`06_1.4bis_node_preflight/`](01_ЧАСТЬ_1_КЛАСТЕР/06_1.4bis_node_preflight): `playbooks/preflight.yml`, `roles/node_preflight/tasks/main.yml`; подготовка сети вынесена в `playbooks/prepare-network.yml` |
| Код установщика Kubespray | [`03_1.2bis_kubespray_anatomy/README.md`](01_ЧАСТЬ_1_КЛАСТЕР/03_1.2bis_kubespray_anatomy/README.md): один официальный клон `v2.31.0`, его `cluster.yml` и роли |
| Наши параметры Kubespray | [`07_1.5_kubespray_full/`](01_ЧАСТЬ_1_КЛАСТЕР/07_1.5_kubespray_full): `inventory/video2/`, `group_vars/k8s_cluster/k8s-cluster.yml`, `group_vars/all/etcd.yml`, `addons-metrics-server.yml` |
| Приложение в своём кластере | [`08_1.6_verify_cluster/app/`](01_ЧАСТЬ_1_КЛАСТЕР/08_1.6_verify_cluster/app): Namespace, Deployment, Service, PDB, debug-клиент и `kustomization.yaml` |
| Сеть между нодами | [`09_1.7_cidr_to_node/`](01_ЧАСТЬ_1_КЛАСТЕР/09_1.7_cidr_to_node): два Pod на разных нодах и Service |
| Сравнение кластеров | [`12_1.10_compare_clusters/`](01_ЧАСТЬ_1_КЛАСТЕР/12_1.10_compare_clusters): общая `base/` и два `overlays/` — `self-managed` и `managed` |

## Если первую часть пропустил

Смотреть её целиком для навигации по этому выпуску необязательно. Но для выполнения
команд нужны доступ в Yandex Cloud, Terraform, `yc`, SSH-ключ и созданные VM.
Начни с [Terraform пяти VM](01_ЧАСТЬ_1_КЛАСТЕР/04_1.3_five_vm_terraform/terraform/README.md).
В его `locals.tf` задаётся состав машин, а `inventory.tf` собирает outputs
`kubespray_inventory`, `kubespray_all_yml` и `external_ha_inventory`.

Если проходил первую часть, используй эти же VM и рабочие файлы. Перед продолжением
проверь, что облачная авторизация ещё действительна, разрешён текущий адрес доступа
и inventory содержит адреса существующих машин. Истёкший токен не требует создавать
VM заново.

Официальный код Kubespray находится в `03_1.2bis_kubespray_anatomy/kubespray/`.
Папка `07_1.5_kubespray_full/` содержит **наши настройки**, а не второй клон.
Рабочий inventory и параметры перед установкой переносятся в `inventory/video2/`
внутри первого клона, как показано в выпуске и тетради.

## Решения учебного стенда

| Машина | Роли |
|---|---|
| `node1` | control plane + etcd; внешний адрес API для ноутбука |
| `node2` | control plane + etcd + worker |
| `node3` | etcd + worker |
| `lb1`, `lb2` | отдельные VM для последующей практики внешнего входа; в Kubernetes не входят |

Параметры в исходниках: Kubespray `v2.31.0`, Kubernetes `1.34.7`, containerd,
Calico, kube-proxy в режиме `iptables`, etcd как host-процесс под systemd.
Это компактный учебный стенд с совмещением ролей. Внешний API через адрес `node1`
не получает автоматическое переключение при отказе этой машины.

Папки `10_1.8_safe_upgrade`, `11_1.9_metallb` и `13_1.11_external_ha` сохранены
в общем репозитории для дальнейшей практики. Их наличие в архиве **не означает**,
что обновление или HA-вход уже выполнены в этом выпуске.

## Тетрадь, бюджет и личные данные

Ссылку на рабочую тетрадь бери из описания выпуска: материалы выдаются через
Telegram-бота. Релиз GitHub фиксирует исходники; PDF не подменяется первой частью.

Облачный стенд платный. Не удаляй VM между разделами, которые используют один
кластер. После завершения нужной практики проверь план удаления в каждом
Terraform-проекте и оставшиеся платные ресурсы в облаке.

Учебные адреса можно показывать при разборе inventory. Токены, приватный SSH-ключ,
данные авторизации из kubeconfig и содержимое state целиком публиковать не нужно.
