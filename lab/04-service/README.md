# Service practice

ClusterIP:

```bash
kubectl apply -f ../03-deployment/web-deployment.yaml
kubectl apply -f web-clusterip.yaml
kubectl get svc
kubectl get endpoints
kubectl get endpointslices
kubectl port-forward svc/web 8080:80
curl localhost:8080
```

NodePort:

```bash
kubectl apply -f web-nodeport.yaml
kubectl get svc web-nodeport
minikube service web-nodeport --url
```

Broken selector:

```bash
kubectl apply -f web-service-broken-selector.yaml
kubectl get endpoints
kubectl apply -f web-clusterip.yaml
kubectl get endpoints
```
