# Requests / limits practice

```bash
kubectl apply -f huge-request-pod.yaml
kubectl get pods
kubectl describe pod huge-request
```

Вывод: scheduler смотрит на requests. Если ресурсов нет — pod Pending.

Рабочий пример:

```bash
kubectl apply -f small-request-pod.yaml
kubectl get pods
```
