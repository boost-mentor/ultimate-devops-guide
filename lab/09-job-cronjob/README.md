# Job / CronJob practice

```bash
kubectl apply -f hello-job.yaml
kubectl get jobs
kubectl get pods
kubectl logs job/hello-job
```

CronJob:

```bash
kubectl apply -f hello-cronjob.yaml
kubectl get cronjob
kubectl get jobs
kubectl delete cronjob hello-cron
```
