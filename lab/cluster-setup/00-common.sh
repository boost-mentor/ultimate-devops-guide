#!/usr/bin/env bash
# Запускать на КАЖДОЙ из 3 нод (master + 2 worker). Ubuntu 22.04/24.04.
# Ставит: containerd (SystemdCgroup) + kubeadm/kubelet/kubectl v1.31, готовит ядро/sysctl.
set -euo pipefail

K8S_MINOR="v1.31"

echo "==> [1/4] swap off"
sudo swapoff -a
sudo sed -i.bak '/\sswap\s/ s/^/#/' /etc/fstab || true

echo "==> [2/4] kernel modules + sysctl"
cat <<EOF | sudo tee /etc/modules-load.d/k8s.conf >/dev/null
overlay
br_netfilter
EOF
sudo modprobe overlay
sudo modprobe br_netfilter
cat <<EOF | sudo tee /etc/sysctl.d/k8s.conf >/dev/null
net.bridge.bridge-nf-call-iptables  = 1
net.bridge.bridge-nf-call-ip6tables = 1
net.ipv4.ip_forward                 = 1
EOF
sudo sysctl --system >/dev/null

echo "==> [3/4] containerd + CRI (SystemdCgroup)"
sudo apt-get update -y
sudo apt-get install -y containerd apt-transport-https ca-certificates curl gpg
sudo mkdir -p /etc/containerd
containerd config default | sudo tee /etc/containerd/config.toml >/dev/null
sudo sed -i 's/SystemdCgroup = false/SystemdCgroup = true/' /etc/containerd/config.toml
sudo systemctl restart containerd
sudo systemctl enable containerd

echo "==> [4/4] kubeadm/kubelet/kubectl ${K8S_MINOR}"
sudo mkdir -p /etc/apt/keyrings
curl -fsSL "https://pkgs.k8s.io/core:/stable:/${K8S_MINOR}/deb/Release.key" \
  | sudo gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg
echo "deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/${K8S_MINOR}/deb/ /" \
  | sudo tee /etc/apt/sources.list.d/kubernetes.list >/dev/null
sudo apt-get update -y
sudo apt-get install -y kubelet kubeadm kubectl
sudo apt-mark hold kubelet kubeadm kubectl
sudo systemctl enable kubelet

echo "✅ COMMON готово на $(hostname). Дальше: master → 10-master.sh, worker → 20-worker.sh"
