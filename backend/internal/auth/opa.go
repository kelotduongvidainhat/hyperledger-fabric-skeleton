package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// OPA URL - Using the service name from docker-compose
const opaURL = "http://opa:8181/v1/data/ownership/authz/allow"

// OPAInput represents the structure expected by our rego policy
type OPAInput struct {
	Input struct {
		User struct {
			Username string `json:"username"`
			Role     string `json:"role"`
			Org      string `json:"org"`
		} `json:"user"`
		Request struct {
			Method string `json:"method"`
			Path   string `json:"path"`
		} `json:"request"`
	} `json:"input"`
}

// OPAResponse represents the result from OPA Data API
type OPAResponse struct {
	Result bool `json:"result"`
}

// CheckAuthorization queries OPA to decide if a request is allowed
func CheckAuthorization(username, role, org, method, path string) (bool, error) {
	input := OPAInput{}
	input.Input.User.Username = username
	input.Input.User.Role = role
	input.Input.User.Org = org
	input.Input.Request.Method = method
	input.Input.Request.Path = path

	jsonData, err := json.Marshal(input)
	if err != nil {
		return false, fmt.Errorf("failed to marshal OPA input: %v", err)
	}

	resp, err := http.Post(opaURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return false, fmt.Errorf("failed to query OPA: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("OPA returned non-200 status: %d", resp.StatusCode)
	}

	var opaResp OPAResponse
	if err := json.NewDecoder(resp.Body).Decode(&opaResp); err != nil {
		return false, fmt.Errorf("failed to decode OPA response: %v", err)
	}

	return opaResp.Result, nil
}
