# ConfigMap / Secret practice

```bash
kubectl apply -f app-config.yaml
kubectl apply -f app-secret.yaml
kubectl get configmap app-config -o yaml
kubectl get secret app-secret -o yaml
```

Декодировать Secret:

```bash
kubectl get secret app-secret -o jsonpath='{.data.DB_PASSWORD}' | base64 -d; echo
```

Подключить как env:

```bash
kubectl apply -f web-with-env.yaml
kubectl logs deploy/web-with-env
```

Вывод: Secret по умолчанию — base64, это не шифрование.
