# Pod practice

```bash
kubectl apply -f nginx-pod.yaml
kubectl get pods
kubectl describe pod nginx-pod
kubectl delete pod nginx-pod
kubectl get pods
```

Вывод: голый Pod после удаления не восстанавливается, если за ним не стоит контроллер.
