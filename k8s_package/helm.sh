#!/bin/sh

echo "helm 설치"
export PATH=/usr/local/bin:$PATH # sudo의 경로를 설정해줘야함.
echo $PATH
curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3
sudo chmod 700 get_helm.sh
bash get_helm.sh

echo "================================================"

# echo "calico 설치"
# helm repo add projectcalico https://docs.tigera.io/calico/charts
# kubectl create namespace tigera-operator
# helm install calico projectcalico/tigera-operator --version v3.26.0 --namespace tigera-operator


