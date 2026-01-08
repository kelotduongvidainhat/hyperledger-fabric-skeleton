package fabric

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
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
	"strings"
)

// CAConfig holds CA connection details
type CAConfig struct {
	URL           string
	MSPID         string
	WalletPath    string
	AdminPath     string
	CAName        string
	ContainerName string
}

// EnrollAdmin enrolls an admin user and saves it to the wallet
func EnrollAdmin(cfg CAConfig, username, secret string) error {
	log.Printf("Enrolling admin %s for %s...", username, cfg.MSPID)
	return EnrollUser(cfg, username, secret)
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

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	
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
	
	var caResp struct {
		Success bool `json:"success"`
		Result  struct {
			Cert string `json:"Cert"` 
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

	certPEMStr := caResp.Result.Cert
	if certPEMStr == "" {
		certPEMStr = caResp.Result.Base64
	}

	certPEM, err := base64.StdEncoding.DecodeString(certPEMStr)
	if err != nil || len(certPEM) == 0 {
		certPEM = []byte(certPEMStr)
	}
	
	if len(certPEM) == 0 {
		return fmt.Errorf("received empty certificate from CA")
	}

	// 6. Encode Private Key to PEM
	x509Encoded, _ := x509.MarshalECPrivateKey(privateKey)
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: x509Encoded})

	// 7. Save to Wallet
	return SaveIdentity(username, certPEM, keyPEM, cfg.WalletPath, cfg.MSPID)
}

// RegisterUser calls CLI to register user
func RegisterUser(cfg CAConfig, username, secret string) (string, error) {
	// Internal TLS cert path within the container (auto-generated)
	caCert := "/etc/hyperledger/fabric-ca-server/tls-cert.pem"

	// Step 1: Enroll Admin locally (in container) to get signing certs
	// This ensures the local CLI has the necessary identities to perform registration
	// Check if already enrolled first to reduce volume of logs
	checkCmd := exec.Command("docker", "exec", cfg.ContainerName, "ls", "/etc/hyperledger/fabric-ca-server/msp/signcerts/cert.pem")
	if err := checkCmd.Run(); err != nil {
		enrollCmd := exec.Command("docker", "exec", cfg.ContainerName, "fabric-ca-client", "enroll",
			"-u", "https://admin:adminpw@localhost:"+strings.Split(cfg.URL, ":")[2],
			"--tls.certfiles", caCert,
		)
		if out, err := enrollCmd.CombinedOutput(); err != nil {
			log.Printf("Warning: Admin local enroll failed for %s: %s", cfg.MSPID, string(out))
		}
	}

	// Step 2: Register the new user
	cmd := exec.Command("docker", "exec", cfg.ContainerName, "fabric-ca-client", "register",
		"--caname", cfg.CAName,
		"--id.name", username,
		"--id.secret", secret,
		"--id.type", "client",
		"-u", "https://localhost:"+strings.Split(cfg.URL, ":")[2],
		"--tls.certfiles", caCert,
	)
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		if bytes.Contains(output, []byte("is already registered")) {
			return "user already registered", nil
		}
		return "", fmt.Errorf("register failed: %v, output: %s", err, string(output))
	}
	
	return string(output), nil
}

// ListIdentities returns the list of all registered identities from the CA
func ListIdentities(cfg CAConfig) (string, error) {
	caCert := "/etc/hyperledger/fabric-ca-server/tls-cert.pem"
	
	// Ensure enrolled for CLI
	checkCmd := exec.Command("docker", "exec", cfg.ContainerName, "ls", "/etc/hyperledger/fabric-ca-server/msp/signcerts/cert.pem")
	if err := checkCmd.Run(); err != nil {
		exec.Command("docker", "exec", cfg.ContainerName, "fabric-ca-client", "enroll",
			"-u", "https://admin:adminpw@localhost:"+strings.Split(cfg.URL, ":")[2],
			"--tls.certfiles", caCert,
		).Run()
	}

	cmd := exec.Command("docker", "exec", cfg.ContainerName, "fabric-ca-client", "identity", "list",
		"-u", "https://admin:adminpw@localhost:"+strings.Split(cfg.URL, ":")[2],
		"--tls.certfiles", caCert,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to list identities: %v, output: %s", err, string(output))
	}

	return string(output), nil
}

