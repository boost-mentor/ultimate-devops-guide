# Calico: доказательство связи между нодами

Файлы открываются строго по номеру:

1. `00-namespace.yaml` — изолирует лабораторную.
2. `10-pod-node2.yaml` — создаёт первый диагностический Pod на `node2`.
3. `20-pod-node3.yaml` — создаёт второй Pod на `node3`.
4. `30-service.yaml` — даёт второму Pod постоянное имя внутри кластера.

В записи сначала разбирается каждый объект, затем весь набор применяется
одной короткой командой `kubectl apply -k .`.

После готовности Pod проверка идёт слоями:

1. `kubectl get pods -n network-proof -o wide` — Pod IP и разные ноды;
2. `kubectl get service,endpointslice -n network-proof` — selector действительно
   связал Service с IP второго Pod;
3. из `calico-a` выполняется `wget -qO- http://calico-b` — DNS-имя Service,
   Service routing и HTTP между node2 и node3;
4. тот же запрос повторяется прямо к Pod IP — доказательство межнодового Pod
   network без Service.

Один успешный ping был бы слабым доказательством: он не проверяет TCP-порт,
Service selector и приложение. Здесь второй Pod действительно слушает 8080, а
Service переводит порт 80 в именованный `targetPort: http`.

Если запрос не проходит, гипотеза «Calico сломан» не принимается первой. Сначала
сверяются Pod readiness, EndpointSlice, DNS resolution и route на обеих нодах.
Пустой EndpointSlice, например, доказывает ошибку selector/readiness, а не
firewall. Recovery принимается только после повторного HTTP 200 с node2 на
node3.
