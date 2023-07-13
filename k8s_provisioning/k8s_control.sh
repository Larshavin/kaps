#!/bin/sh

echo "================================================"

# RHEL/CentOS 7 have reported traffic issues being routed incorrectly due to iptables bypassed
sudo tee /etc/sysctl.d/k8s.conf <<EOF
net.bridge.bridge-nf-call-ip6tables = 1
net.bridge.bridge-nf-call-iptables = 1
EOF
sudo modprobe br_netfilter

echo "================================================"

# 패킷 포워딩하는 옵션 활성화
sudo sysctl net.ipv4.ip_forward=1

echo "================================================"

# set control-plane-ip & k8s-network-cidr by user setting 
# ex) control-plane-ip=10.10.24.121
export control_plane_ip=10.10.24.23
# ex) k8s-network-cidr=172.16.0.0/16 
export k8s_network_cidr=172.16.0.0/16

echo "================================================"

# kubeadm으로 클러스터 및 인증 정보 생성
sudo kubeadm init \
    --control-plane-endpoint "${control_plane_ip}:6443" \
    --pod-network-cidr=${k8s_network_cidr}\
    --upload-certs

echo "================================================"

sudo mkdir -p $HOME/.kube
sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
sudo chown $(id -u):$(id -g) $HOME/.kube/config

echo "================================================"

# helm 설치
export PATH=/usr/local/bin:$PATH # sudo의 경로를 설정해줘야함.
sudo curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3
sudo chmod 700 get_helm.sh
sudo ./get_helm.sh

echo "================================================"

# helm으로 calico CNI 설치 => 설치 경로 패치가 잦은듯?
helm repo add projectcalico https://docs.tigera.io/calico/charts
kubectl create namespace tigera-operator
helm install calico projectcalico/tigera-operator --version v3.26.0 --namespace tigera-operator

echo "================================================"

# install bash-completion for kubectl
sudo dnf install bash-completion -y

echo "================================================"

# kubectl completion on bash-completion dir
kubectl completion bash >/etc/bash_completion.d/kubectl

echo "================================================"

# alias kubectl to k
echo 'alias k=kubectl' >> $HOME/.bashrc
echo 'complete -F __start_kubectl k' >> $HOME/.bashrc