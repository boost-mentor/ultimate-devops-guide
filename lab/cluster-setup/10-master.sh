#!/usr/bin/env bash
# ТОЛЬКО на MASTER (после 00-common.sh). kubeadm init + kubeconfig + Calico + печатает join.
# Использование: ./10-master.sh [ADVERTISE_IP]
#   ADVERTISE_IP — приватный IP мастера, если ноды в одной приватной сети.
#   Без аргумента берётся первый IP (hostname -I) — для публичной сети ок.
set -euo pipefail

CALICO_VER="v3.29.1"
ADVERTISE_IP="${1:-$(hostname -I | awk '{print $1}')}"

echo "==> kubeadm init (advertise ${ADVERTISE_IP}, pod-cidr 192.168.0.0/16)"
sudo kubeadm init \
  --apiserver-advertise-address="${ADVERTISE_IP}" \
  --pod-network-cidr=192.168.0.0/16

echo "==> kubeconfig для текущего пользователя"
mkdir -p "$HOME/.kube"
sudo cp -f /etc/kubernetes/admin.conf "$HOME/.kube/config"
sudo chown "$(id -u):$(id -g)" "$HOME/.kube/config"

echo "==> Calico ${CALICO_VER} (pod CIDR совпадает с --pod-network-cidr)"
kubectl apply -f "https://raw.githubusercontent.com/projectcalico/calico/${CALICO_VER}/manifests/calico.yaml"

echo ""
echo "================ JOIN-КОМАНДА ДЛЯ ВОРКЕРОВ (скопируй целиком) ================"
sudo kubeadm token create --print-join-command
echo "=============================================================================="
echo ""
echo "Дальше: на каждом воркере выполни эту строку через sudo (или ./20-worker.sh '<строка>')."
echo "Через 1–2 мин:  kubectl get nodes   (master станет Ready после старта Calico)"
