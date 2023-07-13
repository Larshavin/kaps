#!/bin/sh

# # For control plane 
# sudo firewall-cmd --add-port={80,443,6443,2379,2380,10250,10251,10252,30000-32767}/tcp --permanent
# sudo firewall-cmd --reload

# kaps queries token & hash to installed control plane. 
export token=  
export hash=

# config for data_plane_nodes only
kubeadm join --token "${token}" \
             --discovery-token-ca-cert-hash "${hash}"