# Final exercise

```bash
minikube addons enable ingress

kubectl apply -f final-namespace.yaml
kubectl apply -f final-app-config.yaml
kubectl apply -f final-app-secret.yaml
kubectl apply -f final-deployment.yaml
kubectl apply -f final-service.yaml
kubectl apply -f final-ingress.yaml

kubectl get all -n final-demo
kubectl get endpoints -n final-demo
kubectl get endpointslices -n final-demo
curl -H "Host: final.local" http://$(minikube ip)
```

Проверка self-healing:

```bash
kubectl delete pod -n final-demo -l app=final-web
kubectl get pods -n final-demo -w
```

Очистка:

```bash
kubectl delete namespace final-demo
```
