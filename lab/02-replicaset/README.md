# ReplicaSet practice

```bash
kubectl apply -f nginx-rs.yaml
kubectl get rs
kubectl get pods
kubectl delete pod <pod-name>
kubectl get pods -w
```

Scale через apply:

```bash
kubectl apply -f nginx-rs-scale-3.yaml
kubectl get rs,pods
```

Вывод: ReplicaSet держит нужное количество pod'ов, но руками его обычно не используют — чаще через Deployment.
