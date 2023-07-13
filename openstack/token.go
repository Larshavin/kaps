package openstack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"kaps/types"
	"net/http"
)

func RequestKeystoneToken(identity types.Identity) (string, error) {

	const (
		identityEndpoint = "http://192.168.15.40:5000/v3/auth/tokens"
	)

	requestBody := map[string]interface{}{
		"auth": map[string]interface{}{
			"identity": map[string]interface{}{
				"methods": []string{"password"},
				"password": map[string]interface{}{
					"user": map[string]interface{}{
						"name":     identity.Username,
						"password": identity.Password,
						"domain": map[string]interface{}{
							"name": identity.DomainName,
						},
					},
				},
			},
			"scope": map[string]interface{}{
				"project": map[string]interface{}{
					"id": identity.ProjectId,
				},
			},
		},
	}

	requestBodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", identityEndpoint, bytes.NewBuffer(requestBodyJSON))
	if err != nil {
		return "", err
	}
	setCommonHeaders(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	tokenID := resp.Header.Get("X-Subject-Token")
	if tokenID == "" {
		return "", fmt.Errorf("failed to obtain the Keystone token")
	}

	return tokenID, nil
}

func setCommonHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Go HTTP Client")
}
