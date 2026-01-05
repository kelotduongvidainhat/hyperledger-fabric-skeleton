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
	"io"
	"log"
	"net/http"
	"os/exec"
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

	client := &http.Client{}
	
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call CA: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("enrollment failed (status %d): %s", resp.StatusCode, string(body))
	}

	// 5. Parse Response
	body, _ := io.ReadAll(resp.Body)
	log.Printf("CA Response: %s", string(body))
	
	var caResp struct {
		Success bool `json:"success"`
		Result  struct {
			Cert string `json:"Cert"` // Some versions use Cert
			Base64 string `json:"Base64"`
		} `json:"result"`
		Errors []interface{} `json:"errors"`
	}
	
	if err := json.Unmarshal(body, &caResp); err != nil {
		return fmt.Errorf("failed to decode CA response: %v", err)
	}

	if !caResp.Success {
		return fmt.Errorf("CA returned error: %v", caResp.Errors)
	}

	// Try Cert field first, then Base64
	certPEMStr := caResp.Result.Cert
	if certPEMStr == "" {
		certPEMStr = caResp.Result.Base64
	}

	// Base64 decode the cert
	certPEM, err := base64.StdEncoding.DecodeString(certPEMStr)
	if err != nil || len(certPEM) == 0 {
		// If not base64, maybe it's raw PEM?
		certPEM = []byte(certPEMStr)
	}
	
	if len(certPEM) == 0 {
		return fmt.Errorf("received empty certificate from CA")
	}

	// 6. Encode Private Key to PEM
	x509Encoded, _ := x509.MarshalECPrivateKey(privateKey)
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: x509Encoded})

	// 7. Save to Wallet
	return SaveIdentity(username, certPEM, keyPEM, cfg.WalletPath)
}

// RegisterUser calls CLI to register user (requires Admin already enrolled in CLI or using bootstrap credentials)
func RegisterUser(cfg CAConfig, username, secret string) (string, error) {
	// Step 1: Enroll Admin locally (in container) to get signing certs
	enrollCmd := exec.Command("docker", "exec", "ca_org1", "fabric-ca-client", "enroll",
		"-u", "http://admin:adminpw@localhost:7054",
	)
	if out, err := enrollCmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("admin enroll failed: %v, output: %s", err, string(out))
	}

	// Step 2: Register the new user
	cmd := exec.Command("docker", "exec", "ca_org1", "fabric-ca-client", "register",
		"--caname", "ca-org1",
		"--id.name", username,
		"--id.secret", secret,
		"--id.type", "client",
		"-u", "http://localhost:7054",
	)
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("register failed: %v, output: %s", err, string(output))
	}
	
	return string(output), nil
}
