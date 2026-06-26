# DaemonSet practice

```bash
kubectl apply -f node-agent.yaml
kubectl get ds
kubectl get pods -o wide
```

Вывод: DaemonSet создаёт pod на каждой подходящей node.
