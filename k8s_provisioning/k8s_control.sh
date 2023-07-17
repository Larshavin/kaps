#!/bin/sh

echo "============= k8s control install =============="
export HOME=/home/centos
export os_user_id=1000
echo "check HOME path => " $HOME
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
sudo chown ${os_user_id}:${os_user_id} $HOME/.kube/config

echo "================================================"

sudo mkdir -p /root/.kube
sudo cp -i /etc/kubernetes/admin.conf /root/.kube/config
sudo chown ${os_user_id}:${os_user_id} /root/.kube/config

echo "================================================"

# install bash-completion for kubectl
sudo dnf install bash-completion -y

echo "================================================"

# kubectl completion on bash-completion dir
kubectl completion bash >/etc/bash_completion.d/kubectl

echo "================================================"

# alias kubectl to k
echo 'alias k=kubectl' >> $HOME/.bashrc
echo 'alias k=kubectl' >> /root/.bashrc
echo 'complete -F __start_kubectl k' >> $HOME/.bashrc
echo 'complete -F __start_kubectl k' >> /root/.bashrc
source $HOME/.bashrc
source /root/.bashrc