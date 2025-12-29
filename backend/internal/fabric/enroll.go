package fabric

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/hyperledger/fabric-gateway/pkg/identity"
)

// EnrollmentStore handles identity storage on disk
type EnrollmentStore struct {
	BaseDir string
}

func NewEnrollmentStore(baseDir string) *EnrollmentStore {
	return &EnrollmentStore{BaseDir: baseDir}
}

// GetIdentity returns an X509Identity from the store
func (s *EnrollmentStore) GetIdentity(label string, mspID string) (*identity.X509Identity, error) {
	certPath := filepath.Join(s.BaseDir, label, "cert.pem")
	certPEM, err := ioutil.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate: %w", err)
	}

	cert, err := identity.CertificateFromPEM(certPEM)
	if err != nil {
		return nil, err
	}

	return identity.NewX509Identity(mspID, cert)
}

// GetSigner returns a signing function from the store
func (s *EnrollmentStore) GetSigner(label string) (identity.Sign, error) {
	keyDir := filepath.Join(s.BaseDir, label, "keystore")
	files, err := ioutil.ReadDir(keyDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read keystore: %w", err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no key files in %s", keyDir)
	}

	keyPath := filepath.Join(keyDir, files[0].Name())
	keyPEM, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read key: %w", err)
	}

	privateKey, err := identity.PrivateKeyFromPEM(keyPEM)
	if err != nil {
		return nil, err
	}

	return identity.NewPrivateKeySign(privateKey)
}

// SaveIdentity saves cert and key to the store
func (s *EnrollmentStore) SaveIdentity(label string, certPEM []byte, keyPEM []byte) error {
	dir := filepath.Join(s.BaseDir, label)
	keyDir := filepath.Join(dir, "keystore")

	if err := os.MkdirAll(keyDir, 0700); err != nil {
		return err
	}

	if err := ioutil.WriteFile(filepath.Join(dir, "cert.pem"), certPEM, 0600); err != nil {
		return err
	}

	if err := ioutil.WriteFile(filepath.Join(keyDir, "key.pem"), keyPEM, 0600); err != nil {
		return err
	}

	return nil
}
