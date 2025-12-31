package fabric

import (
	"fmt"
	"net/url"
	"path/filepath"

	"github.com/hyperledger/fabric-ca/api"
	"github.com/hyperledger/fabric-ca/lib"
	"github.com/hyperledger/fabric-ca/lib/tls"
)

// CAClient wrapper for Fabric CA operations
type CAClient struct {
	Client *lib.Client
	HomeDir string
}

// NewCAClient creates a new CA Client
func NewCAClient(caURL string, homeDir string, tlsCertPath string) (*CAClient, error) {
	_, err := url.Parse(caURL)
	if err != nil {
		return nil, fmt.Errorf("invalid CA URL: %w", err)
	}

	c := &lib.Client{
		HomeDir: homeDir,
		Config:  &lib.ClientConfig{
			URL: caURL,
			MSPDir: filepath.Join(homeDir, "msp"),
			TLS: tls.ClientTLSConfig{
				Enabled:   true,
				CertFiles: []string{tlsCertPath},
			},
		},
	}
	
	// Initialize the client (loads config)
	if err := c.Init(); err != nil {
		return nil, fmt.Errorf("failed to init CA client: %w", err)
	}

	return &CAClient{
		Client:  c,
		HomeDir: homeDir,
	}, nil
}

// Enroll admin or user
func (c *CAClient) Enroll(username, password string) (*lib.Identity, error) {
	enrollmentResponse, err := c.Client.Enroll(&api.EnrollmentRequest{
		Name:   username,
		Secret: password,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to enroll user %s: %w", username, err)
	}
	
	return enrollmentResponse.Identity, nil
}

// Register a new user (requires registrar identity, usually admin)
func (c *CAClient) Register(registrarUsername, registrarPassword, newUsername, newPassword, userType, affiliation string) (string, error) {
	// 1. Enroll Registrar
	registrarEnrollment, err := c.Client.Enroll(&api.EnrollmentRequest{
		Name:   registrarUsername,
		Secret: registrarPassword,
	})
	if err != nil {
		return "", fmt.Errorf("failed to enroll registrar %s: %w", registrarUsername, err)
	}

	// 2. Create Identity object for the registrar to perform operations
	registrarIdentity := registrarEnrollment.Identity

	// 3. Register the new user
	rr := &api.RegistrationRequest{
		Name:        newUsername,
		Secret:      newPassword,
		Type:        userType,
		Affiliation: affiliation,
	}

	response, err := registrarIdentity.Register(rr)
	if err != nil {
		return "", fmt.Errorf("failed to register user %s: %w", newUsername, err)
	}

	return response.Secret, nil
}
