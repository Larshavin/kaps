package types

import (
	"go.mongodb.org/mongo-driver/mongo"
)

type Identity struct {
	Username   string
	Password   string
	DomainName string
	ProjectId  string
}

type Server struct {
	Name           string                      `json:"name"`
	Image          string                      `json:"imageRef"`
	Flavor         string                      `json:"flavorRef"`
	Networks       []ServerCreateNetwork       `json:"networks"`
	SecurityGroups []ServerCreateSecurityGroup `json:"security_groups"`
	KeyName        string                      `json:"key_name"`
	Script         string                      `json:"user_data"`
}

type ServerCreateNetwork struct {
	UUID    string  `json:"uuid"`
	FixedIP *string `json:"fixed_ip,omitempty"`
}

type ServerCreateSecurityGroup struct {
	Name string `json:"name"`
}

type ServerCreateRequest struct {
	Server Server `json:"server"`
}

type KaasCreateRequest struct {
	Name              string `json:"name"`
	Version           string `json:"version"`
	NetworkId         string `json:"network"`
	KeyName           string `json:"keypair"`
	Flavor            string `json:"flavor"`
	ControlPlaneNodes []Node `json:"control_plane_nodes"`
	DataPlaneNodes    []Node `json:"data_plane_nodes"`
	// Image    string `json:"image"`
	// Script string `json:"user_data"`
}

type Node struct {
	Kind    string `json:"kind"`
	FixedIP string `json:"fixed_ip"`
	Main    *bool  `json:"main,omitempty"`
	Name    string
}

type K8SCluster struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	NetworkId string `json:"network"`
	KeyName   string `json:"keypair"`
	ProjectId string `json:"project_id"`
	Members   []Node `json:"members"`
}

type MongoDB struct {
	URI    string
	Client *mongo.Client
}
