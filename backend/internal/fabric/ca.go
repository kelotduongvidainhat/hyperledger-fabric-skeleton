package fabric

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"
	"crypto/tls"
)

// CAConfig holds CA connection details
type CAConfig struct {
	URL        string
	MSPID      string
	WalletPath string
	AdminPath  string
}

// EnrollUser generates a key/CSR and requests enrollment from the CA
func EnrollUser(cfg CAConfig, username, secret string) error {
	// 1. Generate ECDSA Key (P256)
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return fmt.Errorf("failed to generate key: %v", err)
	}

	// 2. Create Certificate Request (CSR)
	template := x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName: username,
		},
	}
	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &template, privateKey)
	if err != nil {
		return fmt.Errorf("failed to create CSR: %v", err)
	}
	csrPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrBytes})

	// 3. Prepare JSON Request
	reqBody := map[string]interface{}{
		"certificate_request": string(csrPEM),
		"profile":            "tls", // Standard profile
		"label":              "",
	}
	jsonBody, _ := json.Marshal(reqBody)

	// 4. Send Request
	req, err := http.NewRequest("POST", cfg.URL+"/api/v1/enroll", bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}
	req.SetBasicAuth(username, secret)
	req.Header.Set("Content-Type", "application/json")

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // Dev mode for CA
	}
	client := &http.Client{Transport: tr}
	
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call CA: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("enrollment failed (status %d): %s", resp.StatusCode, string(body))
	}

	// 5. Parse Response
	// Response structure: { "result": { "Base64": "..." }, "success": true }
	var caResp struct {
		Success bool `json:"success"`
		Result  struct {
			Base64 string `json:"Base64"` // The cert is inside this field
		} `json:"result"`
		Errors []interface{} `json:"errors"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&caResp); err != nil {
		return fmt.Errorf("failed to decode CA response: %v", err)
	}

	if !caResp.Success {
		return fmt.Errorf("CA returned error: %v", caResp.Errors)
	}

	// Base64 decode the cert
	certPEM, err := base64.StdEncoding.DecodeString(caResp.Result.Base64)
	if err != nil {
		return fmt.Errorf("failed to decode cert base64: %v", err)
	}

	// 6. Encode Private Key to PEM
	x509Encoded, _ := x509.MarshalECPrivateKey(privateKey)
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: x509Encoded})

	// 7. Save to Wallet
	return SaveIdentity(username, certPEM, keyPEM, cfg.WalletPath)
}

// RegisterUser (Stub) - needs Admin Token
func RegisterUser(cfg CAConfig, username, secret string) (string, error) {
	// Not implemented in HTTP mode for now.
	// We assume users (e.g. admin) are already registered or registered via CLI.
	// Login (Enroll) works if already registered.
	return "", nil
}
