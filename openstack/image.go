package openstack

import (
	"encoding/json"
	"net/http"
)

func GetImageList(tokenID string) (map[string]interface{}, error) {

	endpoint := "http://192.168.15.40:9292/v2/images"

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

	var imageListResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&imageListResponse)
	if err != nil {
		return nil, err
	}

	return imageListResponse, nil
}
