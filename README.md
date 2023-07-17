# KAPS.md

Kubernetes Auto Provisioning Service on Echo-e cloud system ( State of PoC )

This is the K8s [Cluster API](https://cluster-api.sigs.k8s.io/) or Openstack [Magnum](https://docs.openstack.org/magnum/latest/) like lightweight custom project that uses minimal processes.


# Progress

![35%](http://progress-bar.dev/35?title=kaps)

:checkered_flag: **Status**

 &ensp; :heavy_check_mark: Using Openstack API without [gopher cloud package](http://gophercloud.io/)

 &ensp; :heavy_check_mark: Create Automatically (only one) Control-Plane & Data-Plane Virtual Machine(VM) on Openstack
 
 &ensp; :heavy_check_mark: Export K8S Cloud access token, hash info from created Control-Plane VM (~ 5min m1.medium size)

 &ensp; :heavy_check_mark: Inject kubeadm's Join cmd to Data-Plane Node VM.

 &ensp; :x: Create LB or API Gateway webserver for access to data plane node (They have no floating IP for external user)

 &ensp; :x: Support Multi Control-Plane Cluster 

 &ensp; :x: Middleware authentication

 &ensp; :x: Manage cluster info by using Database (Postgresql or MongoDB)

 &ensp; :x: Connect to private image repository (Pull images such as calico CNI project ... )

 &ensp; :x: K8S cluster Health Checking 

 &ensp; :heavy_exclamation_mark: Support only Centos8 image now

# Quick Start

1) Install [Golang](https://go.dev/) 1.20 

2) Clone repository

3) Run the Code
```bash
swag init && go run main.go
```
4) Check the Swagger docs in [host]:[port]/swagger/index.html

# Structure

```mermaid
flowchart LR
		subgraph ide1 [Openstack]
		B --> Y([VM generate])
		C --> Y
		D --> Y
		end
		subgraph ide2 [Kubernetes]
		Y --> I{Control Plane}
		Y --> J{Data Plane}
		end
		Z[admin] --> |Set config | A
		A --> K([health checking])
		K --> I
		K --> J
	  A[Kaps] --- |compute service| B[nova]
		A[Kaps] --- |network service| C[neutron]
		A[Kaps] --- |storage service| D[cinder]
		A--> |resource management| X[(Database)]
		X --- K
```


```bash
.
├── README.md
├── docs
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
├── go.mod
├── go.sum
├── k8s_provisioning
│   ├── common.sh
│   ├── k8s_control.sh
│   ├── k8s_control_join.sh
│   ├── k8s_data.sh
│   └── test.sh
├── kaas
│   └── kaas.go
├── main.go
├── nohup.out
├── openstack
│   ├── compute.go
│   ├── identity.go
│   ├── image.go
│   ├── network.go
│   └── token.go
├── routes
│   └── route.go
├── run.sh
├── types
│   └── type.go
└── utils
    ├── file.go
	├── rand.go
    └── ssh.go
```

# Docs
TBU

# Contact Us

Email : syyang@forwiz.com
