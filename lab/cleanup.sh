#!/usr/bin/env bash
set -euo pipefail

kubectl delete namespace dev --ignore-not-found
kubectl delete namespace final-demo --ignore-not-found

kubectl delete pod nginx-pod huge-request small-request --ignore-not-found
kubectl delete rs nginx-rs --ignore-not-found
kubectl delete deploy web hello web-with-env probe-demo --ignore-not-found
kubectl delete svc web web-nodeport web-headless hello probe-demo --ignore-not-found
kubectl delete ingress hello-ingress --ignore-not-found
kubectl delete configmap app-config --ignore-not-found
kubectl delete secret app-secret --ignore-not-found
kubectl delete statefulset web-stateful --ignore-not-found
kubectl delete svc web-stateful --ignore-not-found
kubectl delete daemonset node-agent --ignore-not-found
kubectl delete job hello-job --ignore-not-found
kubectl delete cronjob hello-cron --ignore-not-found

echo "Cleanup completed"
