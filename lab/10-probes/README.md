# Probes practice

Good readiness:

```bash
kubectl apply -f probe-demo-good.yaml
kubectl apply -f probe-service.yaml
kubectl get pods
kubectl get endpoints
```

Broken readiness:

```bash
kubectl apply -f probe-demo-bad-readiness.yaml
kubectl get pods
kubectl describe pod <pod-name>
kubectl get endpoints
```

Вывод: Running != Ready. Service должен отправлять трафик только в Ready pod'ы.
