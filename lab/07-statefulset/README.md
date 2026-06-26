# StatefulSet practice

```bash
kubectl apply -f web-stateful.yaml
kubectl get pods
kubectl delete pod web-stateful-1
kubectl get pods -w
```

Вывод: pod возвращается с тем же стабильным именем.
