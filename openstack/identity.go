package openstack

import (
	"encoding/json"
	"net/http"
)

type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Enabled  bool   `json:"enabled"`
	DomainID string `json:"domain_id"`
	Links    struct {
		Self string `json:"self"`
	} `json:"links"`
}

type UserListResponse struct {
	Users []User `json:"users"`
	Links struct {
		Self     string `json:"self"`
		Previous string `json:"previous"`
		Next     string `json:"next"`
	} `json:"links"`
}

func GetUsersList(tokenID string) ([]User, error) {
	endpoint := "http://192.168.15.40:5000/v3/users"

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

	var userListResponse UserListResponse
	err = json.NewDecoder(resp.Body).Decode(&userListResponse)
	if err != nil {
		return nil, err
	}

	return userListResponse.Users, nil
}
