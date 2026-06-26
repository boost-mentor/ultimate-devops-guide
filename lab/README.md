# Kubernetes course manifests

Готовый набор манифестов для практики в minikube.

## Быстрый старт

```bash
minikube start
kubectl get nodes
```

Дальше применяй файлы по порядку:

```bash
kubectl apply -f 00-namespace/dev-namespace.yaml
kubectl apply -f 01-pod/nginx-pod.yaml
kubectl apply -f 02-replicaset/nginx-rs.yaml
kubectl apply -f 03-deployment/web-deployment.yaml
kubectl apply -f 04-service/web-clusterip.yaml
```

## Очистка всего

```bash
kubectl delete namespace dev --ignore-not-found
kubectl delete -f 01-pod/nginx-pod.yaml --ignore-not-found
kubectl delete -f 02-replicaset/nginx-rs.yaml --ignore-not-found
kubectl delete -f 03-deployment/web-deployment.yaml --ignore-not-found
kubectl delete -f 04-service/web-clusterip.yaml --ignore-not-found
kubectl delete -f 04-service/web-nodeport.yaml --ignore-not-found
kubectl delete -f 05-ingress/hello-ingress.yaml --ignore-not-found
kubectl delete -f 05-ingress/hello-service.yaml --ignore-not-found
kubectl delete -f 05-ingress/hello-deployment.yaml --ignore-not-found
kubectl delete -f 06-config-secret/app-config.yaml --ignore-not-found
kubectl delete -f 06-config-secret/app-secret.yaml --ignore-not-found
kubectl delete -f 06-config-secret/web-with-env.yaml --ignore-not-found
kubectl delete -f 07-statefulset/web-stateful.yaml --ignore-not-found
kubectl delete -f 08-daemonset/node-agent.yaml --ignore-not-found
kubectl delete -f 09-job-cronjob/hello-job.yaml --ignore-not-found
kubectl delete -f 09-job-cronjob/hello-cronjob.yaml --ignore-not-found
kubectl delete -f 10-probes/probe-demo-good.yaml --ignore-not-found
kubectl delete -f 10-probes/probe-demo-bad-readiness.yaml --ignore-not-found
kubectl delete -f 11-resources/huge-request-pod.yaml --ignore-not-found
kubectl delete -f 12-rbac/dev-reader-rbac.yaml --ignore-not-found
```

## Важные команды для демонстрации

```bash
kubectl get pods
kubectl get pods -o wide
kubectl get pods --show-labels
kubectl describe pod <pod-name>
kubectl get deploy,rs,pods
kubectl get svc
kubectl get endpoints
kubectl get endpointslices
kubectl rollout status deployment/web
kubectl rollout history deployment/web
kubectl auth can-i get pods
```
