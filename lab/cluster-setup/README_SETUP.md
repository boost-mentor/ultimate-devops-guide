# Развёртывание учебного кластера — Timeweb (1 master + 2 worker)

> 🟢 **На видео основной путь — kubespray** (наш набор ansible-плейбуков): создаю 3 ВМ в Timeweb прямо в кадре и раскатываю одним прогоном `ansible-playbook … cluster.yml`. Как устроен kubespray — отдельное видео.
> 🔧 **Эти kubeadm-скрипты — быстрый fallback**, если хочется без kubespray (или для локальной репетиции). Результат тот же: 3-нодовый кластер с Calico.

> Цель: за ~15 минут получить РЕАЛЬНЫЙ 3-нодовый кластер для записи видео (control plane виден процессами, Calico, cross-node сеть, реальный NodePort/LoadBalancer-pending). Подними VM → прогони 3 скрипта → готово.
> ⚠️ Сами VM создаёшь ты (твой аккаунт Timeweb + оплата). Я подготовил всё остальное до копи-пейста.

## ШАГ 1 — Создать 3 VM в панели Timeweb (≈5 мин)
- **ОС:** Ubuntu 22.04 (или 24.04).
- **Конфиг (дёшево):** master — 2 vCPU / 4 GB / 40 GB; worker-1 и worker-2 — 2 vCPU / 4 GB / 40 GB. (Можно 2 GB на воркеры, но 4 спокойнее.)
- **SSH-ключ:** добавь свой публичный ключ при создании (чтобы заходить без пароля).
- **Сеть:** по возможности добавь все 3 в одну **приватную сеть** (тогда у нод будут приватные IP `10.0.0.x` — чище для демо). Если лень — публичных IP тоже хватит для учебного стенда.
- Имена: `k8s-master`, `k8s-worker-1`, `k8s-worker-2`.
- ⚠️ Запиши IP всех трёх (и приватные, если есть).

> 💡 Альтернатива — Terraform (провайдер `twc`, токен `TWC_TOKEN`): можно, но для 3 VM панель быстрее и без риска со схемой провайдера. Скажи — выпишу `.tf` отдельно и выверю.

## ШАГ 2 — Закинуть скрипты на ноды
С локальной машины (из папки `cluster-setup/`):
```bash
chmod +x *.sh
for IP in <MASTER_IP> <WORKER1_IP> <WORKER2_IP>; do
  scp 00-common.sh root@$IP:/root/
done
scp 10-master.sh check.sh root@<MASTER_IP>:/root/
scp 20-worker.sh root@<WORKER1_IP>:/root/
scp 20-worker.sh root@<WORKER2_IP>:/root/
```
(юзер `root` — дефолт Timeweb; если другой — поправь)

## ШАГ 3 — Common на ВСЕХ трёх нодах
Зайди на каждую (`ssh root@<IP>`) и:
```bash
./00-common.sh        # containerd + kubeadm/kubelet/kubectl v1.31 (~2-3 мин)
```

## ШАГ 4 — Master
На мастере:
```bash
./10-master.sh <MASTER_PRIVATE_IP>     # без аргумента возьмёт публичный — тоже ок
```
Он сделает `kubeadm init`, настроит `kubectl`, поставит Calico и **напечатает join-команду** — скопируй её.

## ШАГ 5 — Воркеры
На каждом воркере вставь напечатанную строку:
```bash
sudo kubeadm join <MASTER_IP>:6443 --token ... --discovery-token-ca-cert-hash sha256:...
```

## ШАГ 6 — Проверка (на master)
```bash
./check.sh            # ждём 3x Ready + calico/coredns Running (1-2 мин)
kubectl get nodes -o wide
```
Готово — кластер живой. Можно тыкаться и писать видео.

---

## Забрать kubeconfig к себе на ноут (чтобы kubectl с локали)
```bash
scp root@<MASTER_IP>:/etc/kubernetes/admin.conf ~/.kube/config-k8s-course
# если master за публичным IP — сервер в конфиге уже верный; если приватный — поправь server: на публичный IP мастера
KUBECONFIG=~/.kube/config-k8s-course kubectl get nodes
```

## Пути для командной шпаргалки (НЕ minikube, а kubeadm!)
- зайти на ноду: `ssh root@<NODE_IP>` (вместо `minikube ssh`)
- процессы control plane: на мастере `ps aux | grep -E 'kube-apiserver|etcd|kube-scheduler|kube-controller-manager'`
- etcd-сертификаты: **`/etc/kubernetes/pki/etcd/`** (`ca.crt`, `server.crt`, `server.key`)
- имя etcd-пода: `etcd-k8s-master` (по hostname мастера)
- kubeconfig мастера: **`/etc/kubernetes/admin.conf`**
- NodePort: курлишь оба воркера по их IP; `type=LoadBalancer` реально висит `<pending>` (нет MetalLB)
- crictl на ноде: `sudo crictl ps` (рантайм containerd)

## Сброс / пересборка (между дублями)
```bash
# на каждой ноде:
sudo kubeadm reset -f && sudo rm -rf /etc/cni/net.d ~/.kube
# потом снова с ШАГ 4 (master) / ШАГ 5 (worker)
```
Или сделай **snapshot** трёх VM после ШАГ 3 (чистый common) — откатываешься и быстро пересобираешь.

## ⚠️ На запись
- Замазывай реальные IP/токены и биллинг-панель Timeweb.
- Прогони весь путь ОДИН раз заранее (грязный прогон) — поймаешь грабли, они же станут живыми врезками.
- Calico-демо (veth/cali, cross-node) — теперь работает по-настоящему: под на worker-1 пингует под на worker-2 через Calico.
