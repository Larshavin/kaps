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

echo "================================================"

echo "cilium cli 설치"
CILIUM_CLI_VERSION=$(curl -s https://raw.githubusercontent.com/cilium/cilium-cli/master/stable.txt)
CLI_ARCH=amd64
if [ "$(uname -m)" = "aarch64" ]; then CLI_ARCH=arm64; fi
curl -L --fail --remote-name-all https://github.com/cilium/cilium-cli/releases/download/${CILIUM_CLI_VERSION}/cilium-linux-${CLI_ARCH}.tar.gz{,.sha256sum}
sha256sum --check cilium-linux-${CLI_ARCH}.tar.gz.sha256sum
sudo tar xzvfC cilium-linux-${CLI_ARCH}.tar.gz /usr/local/bin
rm cilium-linux-${CLI_ARCH}.tar.gz{,.sha256sum}

echo "================================================"

echo "cilium 설치"
cilium install

