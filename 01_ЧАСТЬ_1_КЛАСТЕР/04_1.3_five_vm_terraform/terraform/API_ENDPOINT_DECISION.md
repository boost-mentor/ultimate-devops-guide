# Kubernetes API endpoint: честное ограничение Part 1

## Что делает текущий стенд

Terraform создаёт две control-plane-ноды, но локальный kubeconfig Kubespray в
Part 1 использует разрешённый внешний адрес `node1`. Это позволяет проверить
кластер с ноутбука, однако сам клиентский endpoint остаётся одиночным: остановка
`node1` оборвёт этот конкретный путь к API, даже если второй API server жив.

## Почему это не маскируется словом HA

Отказоустойчивый control plane требует стабильного адреса перед API servers:
cloud/network load balancer, VIP в реально общей L2-сети или routed/BGP address.
Адрес должен попадать в SAN сертификата API server, иметь проверяемый маршрут от
клиента и корректные health checks на `/readyz`.

Пара `lb1`/`lb2` в финальной сцене Part 1 балансирует HTTP приложения через
NodePort. Она не подменяет endpoint Kubernetes API и не доказывает его HA.

## Production gate

До переноса этого решения в production отдельно согласуют владельца DNS/VIP,
маршрут и allowlist, интерфейс VIP, health check, сертификаты и recovery. Гипотезу
«мешает firewall» нельзя принимать без проверки DNS, route и TCP/HTTPS с
timestamp: одинаковый `403` или timeout может приходить от другого слоя.

Временный доступ через разрешённый адрес `node1` допустим для записи, но это
workaround с явным долгом: спроектировать и проверить стабильный API endpoint.
