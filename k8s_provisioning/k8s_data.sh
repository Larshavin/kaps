#!/bin/sh

# # For data plane 


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

# kaps queries token & hash to installed control plane. 
export control_plane_ip=10.10.24.23
export token=rqyt6l.y84nc6lhjnb99ki0
export hash=sha256:3d2386fa6c6382cf11203655238a800737ab439d76f1dd2f3970573c94f2c5eb

echo "================================================"

# config for data_plane_nodes only
sudo kubeadm join  "${control_plane_ip}:6443" \
             --token "${token}" \
             --discovery-token-ca-cert-hash "${hash}"