#!/bin/sh
sleep 30
echo "This is common resource installation script for all k8s node"

# 컨트롤플레인 상태 최신화.
sudo dnf -y upgrade
sudo dnf -y update

# vim configuration
echo 'alias vi=vim' >> /etc/profile

# podman 설치
sudo dnf install -y podman
sudo dnf -y update

# swapoff -a to disable swapping
sudo swapoff -a

# sed to comment the swap partition in /etc/fstab
sudo sed -i.bak -r 's/(.+ swap .+)/#\1/' /etc/fstab

# 패킷 포워딩하는 옵션 활성화
sudo sysctl net.ipv4.ip_forward=1

# #firewalld 설치 및 실행
# sudo dnf install -y firewalld
# sudo systemctl enable --now firewalld

#firewalld 비활성
sudo systemctl stop firewalld
sudo systemctl disable firewalld

# kubernetes repo
sudo cat <<EOF | sudo tee /etc/yum.repos.d/kubernetes.repo
[kubernetes]
name=Kubernetes
baseurl=https://packages.cloud.google.com/yum/repos/kubernetes-el7-\$basearch
enabled=1
gpgcheck=1
repo_gpgcheck=1
gpgkey=https://packages.cloud.google.com/yum/doc/yum-key.gpg https://packages.cloud.google.com/yum/doc/rpm-package-key.gpg
exclude=kubelet kubeadm kubectl
EOF

# Set SELinux in permissive mode (effectively disabling it)
sudo setenforce 0
sudo sed -i 's/^SELINUX=enforcing$/SELINUX=permissive/' /etc/selinux/config

# RHEL/CentOS 7 have reported traffic issues being routed incorrectly due to iptables bypassed
sudo tee /etc/sysctl.d/k8s.conf <<EOF
net.bridge.bridge-nf-call-ip6tables = 1
net.bridge.bridge-nf-call-iptables = 1
EOF
sudo modprobe br_netfilter

# install packages
sudo dnf install epel-release -y
sudo dnf install vim-enhanced -y
sudo dnf install git -y

#install cri-o
export OS=CentOS_8
export VERSION=1.27 #쿠버네티스 하위버전과 같아야 함.
sudo curl -L -o /etc/yum.repos.d/devel:kubic:libcontainers:stable.repo https://download.opensuse.org/repositories/devel:/kubic:/libcontainers:/stable/$OS/devel:kubic:libcontainers:stable.repo
sudo curl -L -o /etc/yum.repos.d/devel:kubic:libcontainers:stable:cri-o:$VERSION.repo https://download.opensuse.org/repositories/devel:kubic:libcontainers:stable:cri-o:$VERSION/$OS/devel:kubic:libcontainers:stable:cri-o:$VERSION.repo
sudo dnf install cri-o -y
sudo systemctl daemon-reload
sudo systemctl enable crio --now

# install kubernetes cluster
sudo dnf install -y kubelet kubeadm kubectl --disableexcludes=kubernetes
sudo systemctl enable --now kubelet