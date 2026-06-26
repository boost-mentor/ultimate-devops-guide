# Слайд 18 — Ingress: один вход на наши сервисы

**Идея за 10 секунд:** `Ingress` — это *правило* маршрутизации HTTP по имени хоста/пути.
Само правило ничего не делает — его **исполняет Ingress Controller** (мы ставим
`ingress-nginx`): он слушает на нодах на :80/:443 и по заголовку `Host:` отправляет
трафик в нужный `Service`.

Маршрутизируем на **наши** сервисы (name-based virtual hosting):

| Запрос | Куда идёт | Что это |
|---|---|---|
| `http://scott.local` | `scott-clusterip` → под `scott-pilgrim` (:8080) | наш Go-сервис Scott Pilgrim Quotes |
| `http://web.local`   | `web` → 10 подов nginx | сервис из урока про Deployment/Service |

---

## 1. Поставить Ingress Controller (один раз на кластер)

> На реальном bare-metal кластере (наш kubespray-стенд) **облачного LoadBalancer нет**,
> поэтому ставим `ingress-nginx` и слушаем прямо на ноде (hostNetwork, :80).

```bash
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.13.0/deploy/static/provider/baremetal/deploy.yaml

# слушать на :80 ноды (а не на случайном NodePort) — patch на hostNetwork:
kubectl -n ingress-nginx patch deploy ingress-nginx-controller \
  -p '{"spec":{"template":{"spec":{"hostNetwork":true,"dnsPolicy":"ClusterFirstWithHostNet"}}}}'

kubectl -n ingress-nginx rollout status deploy/ingress-nginx-controller
kubectl get ingressclass        # должен появиться 'nginx'
```

*(minikube-вариант для своего ноута: `minikube addons enable ingress` — контроллер уже встроен.)*

## 2. Применить правило Ingress на наши сервисы

```bash
kubectl apply -f scott-ingress.yaml
kubectl get ingress
# NAME            CLASS   HOSTS                   PORTS
# scott-ingress   nginx   scott.local,web.local   80
```

## 3. Проверить, что один вход разводит трафик на разные сервисы

```bash
# IP любой ноды, где живёт контроллер (наш стенд: node1)
NODE=<NODE_IP>

curl -s -H "Host: scott.local" http://$NODE/ | grep '<title>'   # -> Scott Pilgrim Quotes
curl -s -H "Host: web.local"   http://$NODE/ | grep '<title>'   # -> Welcome to nginx!
```

**Чтобы открыть в браузере по красивому имени** — добавь в `/etc/hosts` на своём компе:

```
<NODE_IP>  scott.local web.local
```

и заходи на `http://scott.local` / `http://web.local` без всяких заголовков.

---

**Вывод слайда:** Ingress — это *правило* (L7-роутинг по Host/path),
Ingress Controller — *компонент*, который это правило исполняет.
Один контроллер на :80 → сколько угодно сервисов за разными именами,
без отдельного NodePort/LoadBalancer на каждый.
