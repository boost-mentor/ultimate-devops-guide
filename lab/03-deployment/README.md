# Deployment practice

```bash
kubectl apply -f web-deployment.yaml
kubectl get deploy,rs,pods
kubectl delete pod <pod-name>
kubectl get pods -w
```

Rolling update:

```bash
kubectl apply -f web-deployment-v126.yaml
kubectl rollout status deployment/web
kubectl rollout history deployment/web
```

Rollback:

```bash
kubectl rollout undo deployment/web
kubectl rollout status deployment/web
```

Scale:

```bash
kubectl apply -f web-deployment-scale-5.yaml
kubectl get pods
```
