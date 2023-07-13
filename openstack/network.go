package openstack

import (
	"encoding/json"
	"net/http"
)

func GetNetworkList(tokenID string) (map[string]interface{}, error) {

	endpoint := "http://192.168.15.40:9696/v2.0/networks"

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

	var networkListResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&networkListResponse)
	if err != nil {
		return nil, err
	}

	return networkListResponse, nil
}
