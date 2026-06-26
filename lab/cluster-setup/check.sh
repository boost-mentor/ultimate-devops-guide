#!/usr/bin/env bash
# На MASTER: быстрая проверка, что кластер собрался.
set -euo pipefail
echo "=== Ноды (ждём 3x Ready: control-plane + 2 worker) ==="
kubectl get nodes -o wide
echo ""
echo "=== Системные поды (calico-node на каждой ноде, coredns Running) ==="
kubectl get pods -A -o wide | grep -E 'calico|coredns|etcd|kube-apiserver|kube-scheduler|kube-controller' || true
echo ""
echo "Если ноды NotReady — подожди 1–2 мин (Calico поднимается). Логи: kubectl -n kube-system get pods"
