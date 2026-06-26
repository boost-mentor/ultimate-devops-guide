#!/usr/bin/env bash
# ТОЛЬКО на WORKER (после 00-common.sh). Присоединяет воркер к кластеру.
# Вариант 1 (проще): просто вставь и выполни строку, которую напечатал 10-master.sh:
#    sudo kubeadm join <MASTER_IP>:6443 --token ... --discovery-token-ca-cert-hash sha256:...
# Вариант 2: ./20-worker.sh "sudo kubeadm join <MASTER_IP>:6443 --token ... --discovery-token-ca-cert-hash sha256:..."
set -euo pipefail

if [ $# -eq 0 ]; then
  echo "Вставь join-команду из вывода 10-master.sh, напр.:"
  echo "  ./20-worker.sh \"sudo kubeadm join 10.0.0.5:6443 --token abcd.efgh --discovery-token-ca-cert-hash sha256:...\""
  exit 1
fi

echo "==> присоединяю $(hostname) к кластеру"
eval "$*"
echo "✅ воркер присоединён. Проверь на master:  kubectl get nodes -o wide"
