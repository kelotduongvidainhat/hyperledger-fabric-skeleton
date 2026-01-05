package fabric

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hyperledger/fabric-gateway/pkg/identity"
)

// SaveIdentity saves the private key and certificate to the file system
func SaveIdentity(username string, cert []byte, key []byte, walletPath string) error {
	userDir := filepath.Join(walletPath, username)
	if err := os.MkdirAll(userDir, 0700); err != nil {
		return fmt.Errorf("failed to create wallet dir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(userDir, "cert.pem"), cert, 0644); err != nil {
		return fmt.Errorf("failed to write cert: %v", err)
	}
	if err := os.WriteFile(filepath.Join(userDir, "key.pem"), key, 0600); err != nil {
		return fmt.Errorf("failed to write key: %v", err)
	}
	
	// Create MSP ID file for completeness
	if err := os.WriteFile(filepath.Join(userDir, "mspid"), []byte("Org1MSP"), 0644); err != nil {
		return fmt.Errorf("failed to write mspid: %v", err) 
	}

	return nil
}

// GetIdentity loads the credentials from the file system and returns an X509Identity
func GetIdentity(username string, walletPath string) (*identity.X509Identity, identity.Sign, error) {
	userDir := filepath.Join(walletPath, username)
	
	certPath := filepath.Join(userDir, "cert.pem")
	keyPath := filepath.Join(userDir, "key.pem")

	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read cert: %v", err)
	}

	// Gateway Identity
	certificate, err := identity.CertificateFromPEM(certPEM)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse certificate: %v", err)
	}

	id, err := identity.NewX509Identity("Org1MSP", certificate)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create X509Identity: %v", err)
	}

	// Signer
	// We need to implement a parser for the private key to return a Sign function
	// Re-using the logic from client.go's loadPrivateKey would be ideal, 
	// but to avoid circular deps or code duplication, let's copy the helper here or export it.
	// For now, I'll inline the key loading logic.
	
	privateKeyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read key: %v", err)
	}
	
	privateKey, err := identity.PrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		// Fallback for SEC1/EC Private Key if standard parser fails
		block, _ := pem.Decode(privateKeyPEM)
		if block != nil && block.Type == "EC PRIVATE KEY" {
			privateKey, err = x509.ParseECPrivateKey(block.Bytes)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to parse EC private key: %v", err)
			}
		} else {
			return nil, nil, fmt.Errorf("failed to parse private key: %v", err)
		}
	}

	sign, err := identity.NewPrivateKeySign(privateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create signer: %v", err)
	}

	return id, sign, nil
}

// IdentityExists checks if a user is in the wallet
func IdentityExists(username string, walletPath string) bool {
	if _, err := os.Stat(filepath.Join(walletPath, username, "cert.pem")); os.IsNotExist(err) {
		return false
	}
	return true
}

// GetMSPDirFromCryptoConfig returns a path to an enrolled user in crypto-config (For Admin)
// Helper to support legacy/admin loading
func GetMSPDirFromCryptoConfig(basePath string) (string, string, error) {
	// Find the key file
	keystore := filepath.Join(basePath, "msp", "keystore")
	files, err := os.ReadDir(keystore)
	if err != nil {
		return "", "", err
	}
	if len(files) == 0 {
		return "", "", fmt.Errorf("no key found in keystore")
	}
	// return key path and cert path
	return filepath.Join(keystore, files[0].Name()), filepath.Join(basePath, "msp", "signcerts", filepath.Base(basePath)+"-cert.pem"), nil 
}
