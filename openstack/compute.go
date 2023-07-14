package openstack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"kaps/types"
	"net/http"
)

const computeEndpoint string = "http://192.168.15.40:8774/v2.1/"

func GetServersList(tokenID string) (map[string]interface{}, error) {

	// endpoint := "http://192.168.15.40:8774/v2.1/a8e14f37c44d460788ce9fa9d825b5c9/servers"
	endpoint := computeEndpoint + "servers"

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	setCommonHeaders(req)
	req.Header.Set("X-Auth-Token", tokenID)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var serverListResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&serverListResponse)
	if err != nil {
		return nil, err
	}
	return serverListResponse, nil
}

type ServerDetailResponse struct {
	Server struct {
		Addresses        map[string][]Address `json:"addresses"`
		Name             string               `json:"name"`
		ID               string               `json:"id"`
		Description      string               `json:"description"`
		TenantID         string               `json:"tenant_id"`
		Status           string               `json:"status"`
		Locked           bool                 `json:"locked"`
		AvailabilityZone string               `json:"OS-EXT-AZ:availability_zone"`
		Created          string               `json:"created"`
		HostName         string               `json:"OS-EXT-SRV-ATTR:host"`
		Flavor           Flavor               `json:"flavor"`
		KeyName          string               `json:"key_name"`
		// Image             string                   `json:"image"`
		VolumeAttachments []map[string]interface{} `json:"os-extended-volumes:volumes_attached"`
		SecurityGroups    []struct {
			Name string `json:"name"`
		} `json:"security_groups"`
	} `json:"server"`
}

type Address struct {
	Addr string `json:"addr"`
}

type Flavor struct {
	OriginalName string `json:"original_name"`
	RAM          int    `json:"ram"`
	CPU          int    `json:"vcpus"`
	Disk         int    `json:"disk"`
}

func GetServerDetail(tokenID string, serverID string) (*ServerDetailResponse, error) {

	endpoint := "http://192.168.15.40:8774/v2.1/servers/" + serverID

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	setCommonHeaders(req)
	req.Header.Set("X-Auth-Token", tokenID)
	req.Header.Set("X-OpenStack-Nova-API-Version", "2.79")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var serverDetailResponse ServerDetailResponse
	err = json.NewDecoder(resp.Body).Decode(&serverDetailResponse)
	if err != nil {
		return nil, err
	}

	return &serverDetailResponse, nil
}

func CreateServer(tokenID, projectID string, serverInfo types.ServerCreateRequest) (map[string]interface{}, error) {

	endpoint := "http://192.168.15.40:8774/v2.1/" + projectID + "/servers"

	// // make serverInfo to io.Reader
	// serverInfoJSON, err := json.Marshal(serverInfo)
	// if err != nil {
	// 	return nil, err
	// }
	fmt.Println(endpoint)

	body, err := json.Marshal(serverInfo)
	if err != nil {
		return nil, err
	}

	body_buffer := bytes.NewBuffer(body)

	req, err := http.NewRequest("POST", endpoint, body_buffer)
	if err != nil {
		return nil, err
	}
	setCommonHeaders(req)
	req.Header.Set("X-Auth-Token", tokenID)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var serverCreationResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&serverCreationResponse)
	if err != nil {
		return nil, err
	}

	return serverCreationResponse, nil
}

func GetKeypairList(tokenID string) (map[string]interface{}, error) {
	endpoint := computeEndpoint + "os-keypairs"
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	setCommonHeaders(req)
	req.Header.Set("X-Auth-Token", tokenID)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var keypairListResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&keypairListResponse)
	if err != nil {
		return nil, err
	}
	return keypairListResponse, nil
}

func CreateKeypair(tokenID string, keyName string) (map[string]interface{}, error) {

	endpoint := computeEndpoint + "os-keypairs"

	keypairInfo := map[string]any{"keypair": map[string]string{"name": keyName, "type": "ssh"}}

	body, err := json.Marshal(keypairInfo)
	if err != nil {
		return nil, err
	}
	body_buffer := bytes.NewBuffer(body)

	req, err := http.NewRequest("POST", endpoint, body_buffer)
	if err != nil {
		return nil, err
	}
	setCommonHeaders(req)
	req.Header.Set("X-Auth-Token", tokenID)
	req.Header.Set("X-OpenStack-Nova-API-Version", "2.79")

	resp, err := http.DefaultClient.Do(req)
	if err != nil && resp.StatusCode != 409 {
		return nil, err
	} else if err != nil && resp.StatusCode == 409 {
		return nil, nil
	}

	defer resp.Body.Close()

	var keypairCreateResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&keypairCreateResponse)
	if err != nil {
		return nil, err
	}

	return keypairCreateResponse, nil
}
