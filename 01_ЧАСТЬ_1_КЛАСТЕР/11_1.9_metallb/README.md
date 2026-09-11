# MetalLB: внешний адрес для self-managed Kubernetes

Сначала в браузере открываются первичные источники:

- официальный репозиторий: <https://github.com/metallb/metallb>;
- source зафиксированного release:
  <https://github.com/metallb/metallb/tree/v0.16.1>;
- установка release: <https://metallb.io/installation/>;
- настройка: <https://metallb.io/configuration/>.

Для записи используется зафиксированный стабильный релиз `v0.16.1`, а не
текущее состояние ветки `main`.

Перед установкой открывается официальный release manifest, а не URL ветки
`main`. После запуска controller и speaker открываются наши три объекта:

1. `00-ip-address-pool.yaml` — зарезервированный пул внешних адресов;
2. `10-l2-advertisement.yaml` — объявление пула в L2-сети;
3. `devops-may-cry-loadbalancer.yaml` — Service приложения.

L2 и BGP не являются двумя флагами одного опыта. L2Advertisement объявляет
адрес через ARP/NDP в общей локальной сети; BGP требует реального router peer,
согласованные ASN, routing policy и проверку маршрута. Поэтому BGP reference
разделён на два самостоятельных объекта:

- `20-bgp-peer-reference.yaml` — сосед и ASN;
- `30-bgp-advertisement-reference.yaml` — какие адреса анонсировать.

Они намеренно не входят в `kustomization.yaml` и в живой лабораторной не
применяются. Успех L2 доказывается не только полем `EXTERNAL-IP`: speaker logs,
EndpointSlice и HTTP 200 должны одновременно подтверждать рабочий путь.
